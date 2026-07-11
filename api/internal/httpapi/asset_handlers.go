package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/arryaanjain/AI_DAY/internal/assets"
	"github.com/arryaanjain/AI_DAY/internal/auth"
	"net/http"
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
