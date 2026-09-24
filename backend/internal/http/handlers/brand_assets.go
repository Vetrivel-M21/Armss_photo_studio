package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

var validBrandAssetKinds = map[string]bool{
	mysql.BrandAssetLogo:    true,
	mysql.BrandAssetTagline: true,
	mysql.BrandAssetFooter:  true,
}

type BrandAssetsHandler struct {
	BrandAssets *mysql.BrandAssetsRepo
	Store       *storage.Store
	// OnReplace is called after a successful replace so the compositor's
	// in-process asset cache can be invalidated immediately (spec §6: admin
	// replacements must take effect right away, not after a restart).
	OnReplace func()
}

func NewBrandAssetsHandler(brandAssets *mysql.BrandAssetsRepo, store *storage.Store, onReplace func()) *BrandAssetsHandler {
	return &BrandAssetsHandler{BrandAssets: brandAssets, Store: store, OnReplace: onReplace}
}

func (h *BrandAssetsHandler) List(w http.ResponseWriter, r *http.Request) {
	assets, err := h.BrandAssets.ListActive(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load brand assets right now.")
		return
	}
	out := make([]dto.BrandAssetResponse, 0, len(assets))
	for _, a := range assets {
		url, err := h.Store.SignedURL(r.Context(), a.ObjectKey, 15*time.Minute)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not load brand assets right now.")
			return
		}
		out = append(out, dto.BrandAssetResponse{
			ID: a.ID, Kind: a.Kind, ImageURL: url, Meta: a.Meta, Version: a.Version, IsActive: a.IsActive, CreatedAt: a.CreatedAt,
		})
	}
	respond.JSON(w, http.StatusOK, out)
}

// Put implements admin: replace (spec §9's PUT /brand-assets). Multipart:
// "kind" (logo|tagline|footer) and "file". Brand assets are versioned — this
// always adds a new active version rather than mutating in place (spec §10).
func (h *BrandAssetsHandler) Put(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		respond.Error(w, http.StatusBadRequest, "Upload is too large. The limit is 10MB.")
		return
	}
	kind := r.FormValue("kind")
	if !validBrandAssetKinds[kind] {
		respond.Error(w, http.StatusBadRequest, "kind must be one of: logo, tagline, footer.")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Attach the asset file under the \"file\" field.")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil || len(data) > maxUploadBytes {
		respond.Error(w, http.StatusBadRequest, "Upload is too large. The limit is 10MB.")
		return
	}
	contentType := storage.SniffContentType(data)
	if contentType != "image/png" {
		respond.Error(w, http.StatusUnsupportedMediaType, "Brand assets must be PNG files with alpha transparency.")
		return
	}

	objectKey := fmt.Sprintf("brand-assets/%s/%s", kind, uuid.NewString())
	if err := h.Store.Put(r.Context(), objectKey, data, contentType); err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not store the brand asset right now.")
		return
	}

	asset, err := h.BrandAssets.Replace(r.Context(), kind, objectKey, nil)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not save the brand asset right now.")
		return
	}
	if h.OnReplace != nil {
		h.OnReplace()
	}

	url, err := h.Store.SignedURL(r.Context(), asset.ObjectKey, 15*time.Minute)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Brand asset was saved but its preview could not be loaded.")
		return
	}
	respond.JSON(w, http.StatusOK, dto.BrandAssetResponse{
		ID: asset.ID, Kind: asset.Kind, ImageURL: url, Meta: asset.Meta, Version: asset.Version, IsActive: asset.IsActive, CreatedAt: asset.CreatedAt,
	})
}
