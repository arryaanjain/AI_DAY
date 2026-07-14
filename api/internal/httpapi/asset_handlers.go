package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/arryaanjain/AI_DAY/internal/assets"
	"github.com/arryaanjain/AI_DAY/internal/auth"
	"github.com/arryaanjain/AI_DAY/internal/storage"
	"github.com/go-chi/chi/v5"
)

type uploadURLRequest struct {
	AssetType string `json:"assetType"`
	Filename  string `json:"filename"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}

func uploadURLHandler(authService *auth.Service, assetService *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, authService)
		if !ok {
			return
		}
		var request uploadURLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.AssetType != "source_selfie" {
			errorResponse(w, http.StatusBadRequest, "INVALID_ASSET", "A valid source selfie upload is required.")
			return
		}
		intent, err := assetService.CreateUploadIntent(r.Context(), user.ID, request.Filename, request.MimeType, request.SizeBytes)
		if errors.Is(err, assets.ErrInvalidAsset) {
			errorResponse(w, http.StatusBadRequest, "UNSUPPORTED_IMAGE", "The selected image is not allowed.")
			return
		}
		if err != nil {
			errorResponse(w, http.StatusServiceUnavailable, "INTERNAL_ERROR", "Unable to prepare upload.")
			return
		}
		respond(w, http.StatusCreated, map[string]string{"assetId": intent.AssetID, "uploadUrl": intent.URL})
	}
}

func downloadAssetHandler(assetService *assets.Service, storageProvider storage.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		assetID := chi.URLParam(r, "id")
		if assetID == "" {
			errorResponse(w, http.StatusBadRequest, "INVALID_ASSET", "Asset ID required.")
			return
		}

		asset, err := assetService.Get(r.Context(), assetID)
		if err != nil {
			errorResponse(w, http.StatusNotFound, "NOT_FOUND", "Asset not found.")
			return
		}

		dlURL, err := storageProvider.PresignDownload(r.Context(), asset.ObjectKey)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to locate asset file.")
			return
		}

		if strings.HasPrefix(dlURL, "file://") {
			filePath := strings.TrimPrefix(dlURL, "file://")
			mimeType := asset.MimeType
			if mimeType == "" {
				if strings.HasSuffix(filePath, ".pdf") {
					mimeType = "application/pdf"
				} else if strings.HasSuffix(filePath, ".png") {
					mimeType = "image/png"
				}
			}
			w.Header().Set("Content-Type", mimeType)
			filename := filepath.Base(filePath)
			if asset.OriginalName != "" {
				filename = asset.OriginalName
			}
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
			http.ServeFile(w, r, filePath)
			return
		}

		http.Redirect(w, r, dlURL, http.StatusFound)
	}
}
