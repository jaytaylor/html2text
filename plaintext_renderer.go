package html2text

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type plaintextRenderer struct{}

func (r *plaintextRenderer) handleElement(ctx *textifyTraverseContext, node *html.Node) error {
	ctx.justClosedDiv = false

	switch node.DataAtom {
	case atom.Br:
		return ctx.emit("\n")

	case atom.H1, atom.H2, atom.H3:
		subCtx := ctx.newSubContext()
		if err := subCtx.traverseChildren(node); err != nil {
			return err
		}

		str := subCtx.buf.String()
		dividerLen := 0
		for _, line := range strings.Split(str, "\n") {
			if lineLen := len([]rune(line)); lineLen-1 > dividerLen {
				dividerLen = lineLen - 1
			}
		}
		var divider string
		if node.DataAtom == atom.H1 {
			divider = strings.Repeat("*", dividerLen)
		} else {
			divider = strings.Repeat("-", dividerLen)
		}

		if node.DataAtom == atom.H3 {
			return ctx.emit("\n\n" + str + "\n" + divider + "\n\n")
		}
		return ctx.emit("\n\n" + divider + "\n" + str + "\n" + divider + "\n\n")

	case atom.Blockquote:
		ctx.blockquoteLevel++
		ctx.prefix = strings.Repeat(">", ctx.blockquoteLevel) + " "
		if err := ctx.emit("\n"); err != nil {
			return err
		}
		if ctx.blockquoteLevel == 1 {
			if err := ctx.emit("\n"); err != nil {
				return err
			}
		}
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}
		ctx.blockquoteLevel--
		ctx.prefix = strings.Repeat(">", ctx.blockquoteLevel)
		if ctx.blockquoteLevel > 0 {
			ctx.prefix += " "
		}
		return ctx.emit("\n\n")

	case atom.Div:
		if ctx.lineLength > 0 {
			if err := ctx.emit("\n"); err != nil {
				return err
			}
		}
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}
		var err error
		if !ctx.justClosedDiv {
			err = ctx.emit("\n")
		}
		ctx.justClosedDiv = true
		return err

	case atom.Li:
		if err := ctx.emit("* "); err != nil {
			return err
		}

		if err := ctx.traverseChildren(node); err != nil {
			return err
		}

		return ctx.emit("\n")

	case atom.B, atom.Strong:
		subCtx := ctx.newSubContext()
		subCtx.endsWithSpace = true
		if err := subCtx.traverseChildren(node); err != nil {
			return err
		}
		str := subCtx.buf.String()
		return ctx.emit("*" + str + "*")

	case atom.A:
		linkText := ""
		// For simple link element content with single text node only, peek at the link text.
		if node.FirstChild != nil && node.FirstChild.NextSibling == nil && node.FirstChild.Type == html.TextNode {
			linkText = node.FirstChild.Data
		}

		// If image is the only child, take its alt text as the link text.
		if img := node.FirstChild; img != nil && node.LastChild == img && img.DataAtom == atom.Img {
			if altText := getAttrVal(img, "alt"); altText != "" {
				if err := ctx.emit(altText); err != nil {
					return err
				}
			}
		} else if err := ctx.traverseChildren(node); err != nil {
			return err
		}

		hrefLink := ""
		if attrVal := getAttrVal(node, "href"); attrVal != "" {
			attrVal = ctx.normalizeHrefLink(attrVal)
			// Don't print link href if it matches link element content or if the link is empty.
			if !ctx.options.OmitLinks && attrVal != "" && linkText != attrVal {
				hrefLink = "( " + attrVal + " )"
			}
		}

		return ctx.emit(hrefLink)

	case atom.P, atom.Ul:
		return ctx.paragraphHandler(node)

	case atom.Table, atom.Tfoot, atom.Th, atom.Tr, atom.Td:
		if ctx.options.PrettyTables {
			return ctx.handleTableElement(node)
		} else if node.DataAtom == atom.Table {
			return ctx.paragraphHandler(node)
		}
		return ctx.traverseChildren(node)

	case atom.Pre:
		ctx.isPre = true
		err := ctx.traverseChildren(node)
		ctx.isPre = false
		return err

	case atom.Style, atom.Script, atom.Head:
		// Ignore the subtree.
		return nil

	default:
		return ctx.traverseChildren(node)
	}
}
