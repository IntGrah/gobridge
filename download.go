package main

import (
	"mime"
)

func MimeTypeToExtension(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "audio/ogg; codecs=opus":
		fallthrough
	case "audio/ogg":
		return ".ogg"
	case "audio/amr":
		return ".amr"
	case "audio/3gp":
		return ".3gp"
	case "audio/aac":
		return ".aac"
	case "audio/mpeg":
		return ".mp3"
	case "application/pdf":
		return ".pdf"
	case "video/mp4":
		return ".mp4"
	case "text/x-vcard":
		return ".vcf"
	case "text/plain":
		return ".txt"
	case "text/comma-separated-values":
		return ".csv"
	default:
		exts, err := mime.ExtensionsByType(mimeType)
		if err != nil {
			return ""
		}
		return exts[0]
	}
}
