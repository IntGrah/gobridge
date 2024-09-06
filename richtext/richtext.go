package richtext

import (
	"fmt"

	formatting "github.com/delthas/discord-formatting"
)

func Marin() {
	parser := formatting.NewParser(nil)
	ast := parser.Parse("*hi* @everyone <:smile:12345> __what__ **is** `up`?")
	formatting.Walk(ast, func(n formatting.Node, entering bool) {
		switch nn := n.(type) {
		case *formatting.TextNode:
			if entering {
				fmt.Print(nn.Content)
			}
		}
	})
	fmt.Println(formatting.Debug(ast))
}
