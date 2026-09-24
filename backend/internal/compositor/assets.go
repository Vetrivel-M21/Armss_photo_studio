package compositor

import (
	"context"
	"fmt"
	"sync"

	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

// BrandAssetSet is the decoded (still-encoded-as-PNG-bytes) set of active
// brand files needed to composite a poster. FooterIcon is optional — the
// footer bar itself is always drawn as a solid rect in code (spec §6), the
// icon is placed on top of it if one has been uploaded.
type BrandAssetSet struct {
	Logo       []byte
	Tagline    []byte
	FooterIcon []byte
}

// AssetLoader caches the active brand assets in-process so every composite
// call doesn't re-fetch from object storage (spec §17: cache aggressively).
// Call Invalidate after an admin replaces an asset via PUT /brand-assets.
type AssetLoader struct {
	mu          sync.RWMutex
	cached      *BrandAssetSet
	brandAssets *mysql.BrandAssetsRepo
	store       *storage.Store
}

func NewAssetLoader(brandAssets *mysql.BrandAssetsRepo, store *storage.Store) *AssetLoader {
	return &AssetLoader{brandAssets: brandAssets, store: store}
}

func (l *AssetLoader) Invalidate() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cached = nil
}

func (l *AssetLoader) Load(ctx context.Context) (*BrandAssetSet, error) {
	l.mu.RLock()
	if l.cached != nil {
		defer l.mu.RUnlock()
		return l.cached, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cached != nil { // re-check after acquiring the write lock
		return l.cached, nil
	}

	active, err := l.brandAssets.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active brand assets: %w", err)
	}

	set := &BrandAssetSet{}
	for _, a := range active {
		data, err := l.store.Get(ctx, a.ObjectKey)
		if err != nil {
			return nil, fmt.Errorf("load brand asset %s: %w", a.Kind, err)
		}
		switch a.Kind {
		case mysql.BrandAssetLogo:
			set.Logo = data
		case mysql.BrandAssetTagline:
			set.Tagline = data
		case mysql.BrandAssetFooter:
			set.FooterIcon = data
		}
	}

	l.cached = set
	return set, nil
}
