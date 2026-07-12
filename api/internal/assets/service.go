package assets

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arryaanjain/AI_DAY/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidAsset = errors.New("invalid asset")
	ErrNotOwned     = errors.New("asset not owned")
)

type Service struct {
	db       *pgxpool.Pool
	storage  storage.Provider
	bucket   string
	maxBytes int64
	allowed  map[string]bool
}

func New(db *pgxpool.Pool, provider storage.Provider, bucket string, maxBytes int64, allowed []string) *Service {
	valid := map[string]bool{}
	for _, mime := range allowed {
		valid[mime] = true
	}
	return &Service{db: db, storage: provider, bucket: bucket, maxBytes: maxBytes, allowed: valid}
}

type UploadIntent struct{ AssetID, URL, ObjectKey string }

type Asset struct {
	ID           string
	UserID       string
	ObjectKey    string
	OriginalName string
	MimeType     string
	Size         int64
}

func (s *Service) CreateUploadIntent(ctx context.Context, userID, filename, mime string, size int64) (UploadIntent, error) {
	if filename == "" || size <= 0 || size > s.maxBytes || !s.allowed[mime] {
		return UploadIntent{}, ErrInvalidAsset
	}
	id, err := uuid()
	if err != nil {
		return UploadIntent{}, err
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = extensionFor(mime)
	}
	key := fmt.Sprintf("users/%s/source/%s/original%s", userID, id, ext)
	_, err = s.db.Exec(ctx, `INSERT INTO assets(id,user_id,asset_type,bucket,object_key,original_filename,mime_type,size_bytes,metadata) VALUES($1,$2,'source_selfie',$3,$4,$5,$6,$7,'{"uploadStatus":"pending"}')`, id, userID, s.bucket, key, filepath.Base(filename), mime, size)
	if err != nil {
		return UploadIntent{}, err
	}
	if s.storage == nil {
		return UploadIntent{AssetID: id, ObjectKey: key}, nil
	}
	url, err := s.storage.PresignUpload(ctx, key, mime, size)
	return UploadIntent{AssetID: id, URL: url, ObjectKey: key}, err
}

// Get retrieves an asset by ID.
func (s *Service) Get(ctx context.Context, assetID string) (*Asset, error) {
	var asset Asset
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, object_key, original_filename, mime_type, size_bytes
		FROM assets
		WHERE id = $1
	`, assetID).Scan(&asset.ID, &asset.UserID, &asset.ObjectKey, &asset.OriginalName, &asset.MimeType, &asset.Size)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidAsset
		}
		return nil, err
	}

	return &asset, nil
}

func (s *Service) ConfirmOwnership(ctx context.Context, userID, assetID string) error {
	var found bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM assets WHERE id=$1 AND user_id=$2 AND asset_type='source_selfie')`, assetID, userID).Scan(&found)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotOwned
	}
	return nil
}
func uuid() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
func extensionFor(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	}
	return ""
}

var _ = pgx.ErrNoRows
