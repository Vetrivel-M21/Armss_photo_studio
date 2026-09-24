package openai

// mapAspectRatio maps an ArtDirection.AspectRatio (e.g. "4:5", "1:1") to the
// nearest size gpt-image-1 actually supports, defaulting to "auto" when the
// ratio isn't recognised rather than erroring — the image still generates.
func mapAspectRatio(ratio string) string {
	switch ratio {
	case "1:1":
		return "1024x1024"
	case "3:4", "2:3", "4:5", "9:16":
		return "1024x1536"
	case "4:3", "3:2", "5:4", "16:9":
		return "1536x1024"
	default:
		return "auto"
	}
}
