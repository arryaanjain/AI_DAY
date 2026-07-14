package storage
import "context"
type Provider interface {
	PresignUpload(context.Context, string, string, int64) (string, error)
	PresignDownload(context.Context, string) (string, error)
	Delete(context.Context, string) error
	Put(ctx context.Context, objectKey string, data []byte, contentType string) error
	Get(ctx context.Context, objectKey string) ([]byte, error)
}

