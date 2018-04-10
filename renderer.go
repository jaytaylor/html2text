package html2text

import (
	"golang.org/x/net/html"
)

type renderer interface {
	handleElement(ctx *textifyTraverseContext, node *html.Node) error
}
