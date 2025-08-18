package tool

import (
	"net/url"
	"regexp"
	"strings"
)

func extractTitle(text string) string {
	words := strings.Fields(text)
	if len(words) > 8 {
		return strings.Join(words[:8], " ") + "..."
	}
	return text
}

func extractTextSimple(html string) string {
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	html = scriptRegex.ReplaceAllString(html, "")

	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	html = styleRegex.ReplaceAllString(html, "")

	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text := tagRegex.ReplaceAllString(html, " ")

	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

func extractTitleSimple(html string) string {
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func extractLinksSimple(html string, baseURL *url.URL) []map[string]any {
	var links []map[string]any

	linkRegex := regexp.MustCompile(`(?i)<a[^>]*href\s*=\s*["']([^"']*)["'][^>]*>(.*?)</a>`)
	matches := linkRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 2 {
			href := match[1]
			text := extractTextSimple(match[2])

			if href != "" {
				if linkURL, err := baseURL.Parse(href); err == nil {
					if text == "" {
						text = href
					}
					links = append(links, map[string]any{
						"url":  linkURL.String(),
						"text": text,
					})
				}
			}
		}
	}

	return links
}
