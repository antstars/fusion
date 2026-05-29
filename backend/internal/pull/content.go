package pull

import (
	"io"
	"strings"
	"unicode"

	"github.com/0x2E/fusion/internal/store"
	"golang.org/x/net/html"
)

var readableMediaTags = map[string]struct{}{
	"audio":   {},
	"embed":   {},
	"iframe":  {},
	"img":     {},
	"object":  {},
	"picture": {},
	"source":  {},
	"video":   {},
}

var ignoredTextTags = map[string]struct{}{
	"script":   {},
	"style":    {},
	"template": {},
}

func hasReadableItemContent(content string) bool {
	if !hasVisibleText(content) && strings.TrimSpace(content) == "" {
		return false
	}

	tokenizer := html.NewTokenizer(strings.NewReader(content))
	ignoredTextDepth := 0

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return tokenizer.Err() != io.EOF && hasVisibleText(content)
		case html.TextToken:
			if ignoredTextDepth == 0 && hasVisibleText(string(tokenizer.Text())) {
				return true
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			tagName, _ := tokenizer.TagName()
			tag := strings.ToLower(string(tagName))
			if _, ok := readableMediaTags[tag]; ok {
				return true
			}
			if tokenType == html.StartTagToken {
				if _, ok := ignoredTextTags[tag]; ok {
					ignoredTextDepth++
				}
			}
		case html.EndTagToken:
			tagName, _ := tokenizer.TagName()
			tag := strings.ToLower(string(tagName))
			if _, ok := ignoredTextTags[tag]; ok && ignoredTextDepth > 0 {
				ignoredTextDepth--
			}
		}
	}
}

func buildReadableItemInputs(items []*ParsedItem) ([]store.BatchCreateItemInput, int) {
	inputs := make([]store.BatchCreateItemInput, 0, len(items))
	skipped := 0
	for _, item := range items {
		if !hasReadableItemContent(item.Content) {
			skipped++
			continue
		}

		inputs = append(inputs, store.BatchCreateItemInput{
			GUID:    item.GUID,
			Title:   item.Title,
			Link:    item.Link,
			Content: item.Content,
			PubDate: item.PubDate,
		})
	}

	return inputs, skipped
}

func hasVisibleText(value string) bool {
	text := html.UnescapeString(value)
	return strings.TrimFunc(text, unicode.IsSpace) != ""
}
