package html2text

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
	"github.com/ssor/bom"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Options provide toggles and overrides to control specific rendering behaviors.
type Options struct {
	PrettyTables bool // Turns on pretty ASCII rendering for table elements.
	OmitLinks    bool // Turns on omitting links.
	Markdown     bool // Turns on markdown mode.
}

// FromHTMLNode renders text output from a pre-parsed HTML document.
func FromHTMLNode(doc *html.Node, o ...Options) (string, error) {
	var (
		options  Options
		renderer renderer
	)

	if len(o) > 0 {
		options = o[0]
	}

	if options.Markdown {
		renderer = &markdownRenderer{}
	} else {
		renderer = &plaintextRenderer{}
	}
	ctx := textifyTraverseContext{
		buf:      bytes.Buffer{},
		renderer: renderer,
		options:  options,
	}

	if err := ctx.traverse(doc); err != nil {
		return "", err
	}

	text := strings.Replace(ctx.buf.String(), "\n ", "\n", -1)
	text = newlineRe.ReplaceAllString(text, "\n\n")
	text = strings.TrimSpace(text)
	return text, nil
}

// FromReader renders text output after parsing HTML for the specified
// io.Reader.
func FromReader(reader io.Reader, options ...Options) (string, error) {
	newReader, err := bom.NewReaderWithoutBom(reader)
	if err != nil {
		return "", err
	}
	doc, err := html.Parse(newReader)
	if err != nil {
		return "", err
	}
	return FromHTMLNode(doc, options...)
}

// FromString parses HTML from the input string, then renders the text form.
func FromString(input string, options ...Options) (string, error) {
	bs := bom.CleanBom([]byte(input))
	text, err := FromReader(bytes.NewReader(bs), options...)
	if err != nil {
		return "", err
	}
	return text, nil
}

var (
	spacingRe = regexp.MustCompile(`[ \r\n\t]+`)
	newlineRe = regexp.MustCompile(`\n\n+`)
)

// traverseTableCtx holds text-related context.
type textifyTraverseContext struct {
	buf bytes.Buffer

	renderer        renderer
	prefix          string
	tableCtx        tableTraverseContext
	options         Options
	endsWithSpace   bool
	justClosedDiv   bool
	blockquoteLevel int
	lineLength      int
	isPre           bool
}

// tableTraverseContext holds table ASCII-form related context.
type tableTraverseContext struct {
	header     []string
	body       [][]string
	footer     []string
	tmpRow     int
	isInFooter bool
}

func (tableCtx *tableTraverseContext) init() {
	tableCtx.body = [][]string{}
	tableCtx.header = []string{}
	tableCtx.footer = []string{}
	tableCtx.isInFooter = false
	tableCtx.tmpRow = 0
}

func (ctx *textifyTraverseContext) handleElement(node *html.Node) error {
	if err := ctx.renderer.handleElement(ctx, node); err != nil {
		return err
	}
	return nil
}

// paragraphHandler renders node children surrounded by double newlines.
func (ctx *textifyTraverseContext) paragraphHandler(node *html.Node) error {
	if err := ctx.emit("\n\n"); err != nil {
		return err
	}
	if err := ctx.traverseChildren(node); err != nil {
		return err
	}
	return ctx.emit("\n\n")
}

// handleTableElement is only to be invoked when options.PrettyTables is active.
func (ctx *textifyTraverseContext) handleTableElement(node *html.Node) error {
	if !ctx.options.PrettyTables {
		panic("handleTableElement invoked when PrettyTables not active")
	}

	switch node.DataAtom {
	case atom.Table:
		if err := ctx.emit("\n\n"); err != nil {
			return err
		}

		// Re-intialize all table context.
		ctx.tableCtx.init()

		// Browse children, enriching context with table data.
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}

		buf := &bytes.Buffer{}
		table := tablewriter.NewWriter(buf)
		table.SetHeader(ctx.tableCtx.header)
		table.SetFooter(ctx.tableCtx.footer)
		table.AppendBulk(ctx.tableCtx.body)

		// Render the table using ASCII.
		table.Render()
		if err := ctx.emit(buf.String()); err != nil {
			return err
		}

		return ctx.emit("\n\n")

	case atom.Tfoot:
		ctx.tableCtx.isInFooter = true
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}
		ctx.tableCtx.isInFooter = false

	case atom.Tr:
		ctx.tableCtx.body = append(ctx.tableCtx.body, []string{})
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}
		ctx.tableCtx.tmpRow++

	case atom.Th:
		res, err := ctx.renderEachChild(node)
		if err != nil {
			return err
		}

		ctx.tableCtx.header = append(ctx.tableCtx.header, res)

	case atom.Td:
		res, err := ctx.renderEachChild(node)
		if err != nil {
			return err
		}

		if ctx.tableCtx.isInFooter {
			ctx.tableCtx.footer = append(ctx.tableCtx.footer, res)
		} else {
			ctx.tableCtx.body[ctx.tableCtx.tmpRow] = append(ctx.tableCtx.body[ctx.tableCtx.tmpRow], res)
		}

	}
	return nil
}

func (ctx *textifyTraverseContext) traverse(node *html.Node) error {
	switch node.Type {
	case html.TextNode:
		var data string
		if ctx.isPre {
			data = node.Data
		} else {
			data = strings.Trim(spacingRe.ReplaceAllString(node.Data, " "), " ")
		}
		return ctx.emit(data)

	case html.ElementNode:
		return ctx.handleElement(node)

	default:
		return ctx.traverseChildren(node)
	}
}

func (ctx *textifyTraverseContext) traverseChildren(node *html.Node) error {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if err := ctx.traverse(c); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *textifyTraverseContext) emit(data string) error {
	if data == "" {
		return nil
	}
	var (
		lines = ctx.breakLongLines(data)
		err   error
	)
	for _, line := range lines {
		runes := []rune(line)
		startsWithSpace := unicode.IsSpace(runes[0])
		if !startsWithSpace && !ctx.endsWithSpace && !strings.HasPrefix(data, ".") {
			if err = ctx.buf.WriteByte(' '); err != nil {
				return err
			}
			ctx.lineLength++
		}
		ctx.endsWithSpace = unicode.IsSpace(runes[len(runes)-1])
		for _, c := range line {
			if _, err = ctx.buf.WriteString(string(c)); err != nil {
				return err
			}
			ctx.lineLength++
			if c == '\n' {
				ctx.lineLength = 0
				if ctx.prefix != "" {
					if _, err = ctx.buf.WriteString(ctx.prefix); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

const maxLineLen = 74

func (ctx *textifyTraverseContext) breakLongLines(data string) []string {
	// Only break lines when in blockquotes.
	if ctx.blockquoteLevel == 0 {
		return []string{data}
	}
	var (
		ret      = []string{}
		runes    = []rune(data)
		l        = len(runes)
		existing = ctx.lineLength
	)
	if existing >= maxLineLen {
		ret = append(ret, "\n")
		existing = 0
	}
	for l+existing > maxLineLen {
		i := maxLineLen - existing
		for i >= 0 && !unicode.IsSpace(runes[i]) {
			i--
		}
		if i == -1 {
			// No spaces, so go the other way.
			i = maxLineLen - existing
			for i < l && !unicode.IsSpace(runes[i]) {
				i++
			}
		}
		ret = append(ret, string(runes[:i])+"\n")
		for i < l && unicode.IsSpace(runes[i]) {
			i++
		}
		runes = runes[i:]
		l = len(runes)
		existing = 0
	}
	if len(runes) > 0 {
		ret = append(ret, string(runes))
	}
	return ret
}

func (ctx *textifyTraverseContext) normalizeHrefLink(link string) string {
	link = strings.TrimSpace(link)
	link = strings.TrimPrefix(link, "mailto:")
	return link
}

// renderEachChild visits each direct child of a node and collects the sequence of
// textuual representaitons separated by a single newline.
func (ctx *textifyTraverseContext) renderEachChild(node *html.Node) (string, error) {
	buf := &bytes.Buffer{}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		s, err := FromHTMLNode(c, ctx.options)
		if err != nil {
			return "", err
		}
		if _, err = buf.WriteString(s); err != nil {
			return "", err
		}
		if c.NextSibling != nil {
			if err = buf.WriteByte('\n'); err != nil {
				return "", err
			}
		}
	}
	return buf.String(), nil
}

func (ctx *textifyTraverseContext) newSubContext() *textifyTraverseContext {
	subCtx := &textifyTraverseContext{
		renderer: ctx.renderer,
	}
	return subCtx
}

func getAttrVal(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}
