package main

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type HTMLParser struct {
	config *Config
}

func NewHTMLParser(config *Config) *HTMLParser {
	return &HTMLParser{config: config}
}

func (p *HTMLParser) ExtractLinks(body io.Reader, baseURL string) ([]string, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	var links []string
	var visit func(*html.Node)

	visit = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var attrKey string
			switch n.Data {
			case "a":
				attrKey = "href"
			case "link", "script":
				attrKey = "src"
				if n.Data == "link" {
					attrKey = "href"
				}
			case "img":
				attrKey = "src"
			}

			if attrKey != "" {
				for i, attr := range n.Attr {
					if attr.Key == attrKey {
						link := attr.Val
						if link != "" && !strings.HasPrefix(link, "#") && !strings.HasPrefix(link, "mailto:") {
							absolute, err := p.toAbsoluteURL(link, baseURL)
							if err == nil && p.config.IsSameDomain(absolute) {
								links = append(links, absolute)
							}
						}

						n.Attr[i].Val = p.modifyLink(link, baseURL)
						break
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}

	visit(doc)
	return links, nil
}

func (p *HTMLParser) toAbsoluteURL(link, baseURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	rel, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	absolute := base.ResolveReference(rel)
	return absolute.String(), nil
}

func (p *HTMLParser) modifyLink(link, baseURL string) string {
	// Здесь можно модифицировать ссылки для локального просмотра
	// Например, преобразовывать абсолютные ссылки в относительные
	return link
}
