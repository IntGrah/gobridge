package richtext

import "fmt"

func Format(username, text string) string {
	return fmt.Sprintf("@%s\n%s", username, text)
}

// func FormatQuote(qUsername, qText, username, text string) string {
// 	return fmt.Sprintf("Replying to @%s\n%s\n\n*@%s*\n%s", qUsername, qText, username, text)
// }
