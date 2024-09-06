package bridge

import (
	"fmt"
	"io"
	"mime"
	"net/http"
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
		fmt.Printf("Unknown mimeType: %v\n", mimeType)
		exts, err := mime.ExtensionsByType(mimeType)
		if err != nil {
			return ""
		}
		fmt.Printf("Best guess: %v\n", exts[0])
		return exts[0]
	}
}

func Download(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	return body
}
