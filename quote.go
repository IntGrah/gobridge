package main

import "fmt"

func quoteFormat(username, text string) string {
	return fmt.Sprintf("@%s\n%s", username, text)
}
