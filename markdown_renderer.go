package html2text

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type markdownRenderer struct {
}

func (r *markdownRenderer) handleElement(ctx *textifyTraverseContext, node *html.Node) error {
	ctx.justClosedDiv = false

	switch node.DataAtom {
	case atom.Br:
		return ctx.emit("\n")

	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		subCtx := textifyTraverseContext{}
		if err := subCtx.traverseChildren(node); err != nil {
			return err
		}

		str := subCtx.buf.String()

		var prefix string
		switch node.DataAtom {
		case atom.H1:
			prefix = "#"
		case atom.H2:
			prefix = "##"
		case atom.H3:
			prefix = "###"
		case atom.H4:
			prefix = "####"
		case atom.H5:
			prefix = "#####"
		case atom.H6:
			prefix = "######"
		}

		return ctx.emit("\n\n" + prefix + " " + str + "\n\n")

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
		subCtx := textifyTraverseContext{}
		subCtx.endsWithSpace = true
		if err := subCtx.traverseChildren(node); err != nil {
			return err
		}
		str := subCtx.buf.String()

		if ctx.options.Markdown {
			return ctx.emit("**" + str + "**")
		}
		return ctx.emit("*" + str + "*")

	case atom.A:
		linkText := ""
		// For simple link element content with single text node only, peek at the link text.
		if node.FirstChild != nil && node.FirstChild.NextSibling == nil && node.FirstChild.Type == html.TextNode {
			linkText = node.FirstChild.Data
		}

		if ctx.options.Markdown {
			// If image is the only child, take its alt text as the link text.
			if img := node.FirstChild; img != nil && node.LastChild == img && img.DataAtom == atom.Img {
				if altText := getAttrVal(img, "alt"); altText != "" {

				}
			}
			hrefLink := ""
			if attrVal := getAttrVal(node, "href"); attrVal != "" {
				attrVal = ctx.normalizeHrefLink(attrVal)
				// Don't print link href if it matches link element content or if the link is empty.
				if !ctx.options.OmitLinks && attrVal != "" && linkText != attrVal {
					hrefLink = fmt.Sprintf("[%v](%v)", linkText, attrVal)
				}
			}
			return ctx.emit(hrefLink)
		} else {
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
		}

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
