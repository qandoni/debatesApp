package core_config

import "strings"

const MaxAvatarSize = 5 * 1024 * 1024

func ExtensionByContentType(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
func ExtractObjectName(avatarURL string) string {
	const bucket = "debates-images"

	parts := strings.SplitN(avatarURL, bucket+"/", 2)
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}
