package parser

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

const tokenizerErrTypeEOF = "EOF"

type Parser struct {
	acceptableMimeType map[string]bool
}

func NewParser() *Parser {
	return &Parser{
		acceptableMimeType: map[string]bool{
			"text/html":  true,
			"text/css":   true,
			"image/png":  true,
			"image/jpeg": true,
			"image/gif":  true,
		},
	}
}

func (p *Parser) ParseLinks(body []byte) ([]string, error) {
	links := make([]string, 0)
	tokenizer := html.NewTokenizer(strings.NewReader(string(body)))

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err().Error() != tokenizerErrTypeEOF {
				return links, fmt.Errorf("unable to locate any link")
			}
			return links, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "a" || token.Data == "link" || token.Data == "script" {
				for _, attr := range token.Attr {
					if attr.Key == "href" || attr.Key == "src" {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}
}
