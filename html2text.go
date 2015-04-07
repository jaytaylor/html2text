package html2text

import (
	"bytes"
	"golang.org/x/net/html"
	"io"
	"regexp"
	"strings"
)

var (
	spacingRe = regexp.MustCompile(`[ \r\n\t]+`)
	newlineRe = regexp.MustCompile(`\n\n+`)
)

func textify(node *html.Node, buf *bytes.Buffer) error {
	var err error
	var noRecurse bool
	switch node.Type {
	case html.TextNode:
		data := strings.Trim(spacingRe.ReplaceAllString(node.Data, " "), "\r\n \t")
		if len(data) > 0 {
			if err = buf.WriteByte('\n'); err != nil {
				return err
			}
			if _, err = buf.WriteString(data); err != nil {
				return err
			}
			buf.WriteByte('\n')
		}
	case html.ElementNode:
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				if _, err = buf.WriteString(" " + attr.Val); err != nil {
					return err
				}
				noRecurse = true
				break
			}
		}
	}
	if !noRecurse {
		beforeLen := buf.Len()
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if err = textify(c, buf); err != nil {
				return err
			}
		}
		if afterLen := buf.Len(); beforeLen != afterLen {
			buf.WriteByte('\n')
		}
	}
	return nil
}

func FromReader(reader io.Reader) (string, error) {
	buf := &bytes.Buffer{}
	doc, err := html.Parse(reader)
	if err != nil {
		return "", err
	}
	if err = textify(doc, buf); err != nil {
		return "", err
	}
	text := strings.TrimSpace(newlineRe.ReplaceAllString(strings.Replace(buf.String(), "\n ", "\n", -1), "\n\n"))
	return text, nil
}

func FromString(input string) (string, error) {
	text, err := FromReader(strings.NewReader(input))
	if err != nil {
		return "", err
	}
	return text, nil
}
