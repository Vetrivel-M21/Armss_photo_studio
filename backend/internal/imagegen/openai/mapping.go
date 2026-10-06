package openai

// mapAspectRatio maps an ArtDirection.AspectRatio (e.g. "4:5", "1:1") to the
// exact dimensions supported by DALL-E 3 ("1024x1024", "1024x1792", "1792x1024").
func mapAspectRatio(ratio string) string {
	switch ratio {
	case "1:1":
		return "1024x1024"
	case "3:4", "2:3", "4:5", "9:16":
		return "1024x1792"
	case "4:3", "3:2", "5:4", "16:9":
		return "1792x1024"
	default:
		return "1024x1024"
	}
}
