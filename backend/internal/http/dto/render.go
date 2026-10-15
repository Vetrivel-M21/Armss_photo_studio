package dto

type RenderRequest struct {
	Provider    string `json:"provider,omitempty"`
	AspectRatio string `json:"aspectRatio,omitempty"`
}

// ZoneOverrideRequest mirrors compositor.ZoneOverride but as wire-friendly
// pointers so a client can send only the axes it wants to nudge.
type ZoneOverrideRequest struct {
	X *float64 `json:"x,omitempty"`
	Y *float64 `json:"y,omitempty"`
	W *float64 `json:"w,omitempty"`
	H *float64 `json:"h,omitempty"`
}

type CompositeRequest struct {
	ZoneOverrides map[string]ZoneOverrideRequest `json:"zoneOverrides,omitempty"`
}
