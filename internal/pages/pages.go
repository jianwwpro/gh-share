package pages

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
)

type ShareItem struct {
	Slug  string    `json:"slug"`
	Date  string    `json:"date"`
	Title string    `json:"title"`
	Lede  string    `json:"lede,omitempty"`
	URL   string    `json:"url"`
	Path  string    `json:"path"`
	MTime time.Time `json:"mtime"`
}

type CreateOptions struct {
	Title    string
	Slug     string
	Type     string // note, chat, code, media, raw
	Theme    string // paper, dark, terminal, or empty
	Lede     string
	Tags     []string
	CodeLang string
	CodeFile string
}

// GenerateSlug generates a slug like 2026-11-04-hello-world from a title
func GenerateSlug(title string) string {
	today := time.Now().Format("2006-01-02")
	kebab := ToKebabCase(title)
	if kebab == "" {
		kebab = "share"
	}
	// Limit to ~5 words / 40 chars
	parts := strings.Split(kebab, "-")
	if len(parts) > 5 {
		parts = parts[:5]
	}
	kebab = strings.Join(parts, "-")
	return fmt.Sprintf("%s-%s", today, kebab)
}

func ToKebabCase(s string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}]+`)
	cleaned := reg.ReplaceAllString(s, "-")
	cleaned = strings.ToLower(strings.Trim(cleaned, "-"))
	return cleaned
}

// BuildHTML generates the complete HTML document
func BuildHTML(content string, opts CreateOptions, basePagesURL string) (string, error) {
	trimmed := strings.TrimSpace(content)
	if opts.Type == "raw" || strings.HasPrefix(strings.ToLower(trimmed), "<!doctype html") || strings.HasPrefix(strings.ToLower(trimmed), "<html") {
		return content, nil
	}

	today := time.Now().Format("2006-01-02")
	title := opts.Title
	if title == "" {
		title = "Untitled Share"
	}

	themeAttr := ""
	if opts.Theme != "" {
		themeAttr = fmt.Sprintf(` data-theme="%s"`, html.EscapeString(opts.Theme))
	}

	tagsHTML := ""
	for _, tag := range opts.Tags {
		tagsHTML += fmt.Sprintf(`<span class="tag">%s</span> `, html.EscapeString(tag))
	}

	ledeHTML := ""
	if opts.Lede != "" {
		ledeHTML = fmt.Sprintf(`<p class="lede">%s</p>`, html.EscapeString(opts.Lede))
	}

	var bodyHTML string
	switch opts.Type {
	case "code":
		lang := opts.CodeLang
		if lang == "" {
			lang = "text"
		}
		dataTitle := ""
		if opts.CodeFile != "" && opts.CodeFile != "." {
			dataTitle = fmt.Sprintf(` data-title="%s"`, html.EscapeString(opts.CodeFile))
		}
		escapedCode := html.EscapeString(content)
		bodyHTML = fmt.Sprintf(`<pre%s><code class="language-%s">%s</code></pre>`, dataTitle, html.EscapeString(lang), escapedCode)

	case "chat":
		// If content is already chat divs, use it; otherwise wrap lines or parse
		if strings.Contains(content, `class="msg"`) {
			bodyHTML = fmt.Sprintf(`<section class="chat">%s</section>`, content)
		} else {
			// Convert markdown or plain text to a chat section
			bodyHTML = fmt.Sprintf(`<section class="chat"><div class="msg" data-role="user"><p>%s</p></div></section>`, html.EscapeString(content))
		}

	case "media":
		bodyHTML = content

	default: // "note" or markdown
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(content), &buf); err != nil {
			bodyHTML = fmt.Sprintf("<p>%s</p>", html.EscapeString(content))
		} else {
			bodyHTML = buf.String()
		}
	}

	escapedTitle := html.EscapeString(title)
	escapedDesc := html.EscapeString(opts.Lede)
	if escapedDesc == "" {
		escapedDesc = escapedTitle
	}

	doc := fmt.Sprintf(`<!doctype html>
<html lang="zh-CN"%s>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="robots" content="noindex, nofollow">
  <title>%s</title>
  <meta name="description" content="%s">
  <meta property="og:title" content="%s">
  <meta property="og:description" content="%s">
  <link rel="stylesheet" href="../../assets/share.css">
  <script src="../../assets/share.js" defer></script>
</head>
<body>
  <main class="page">
    <header class="meta">
      <h1>%s</h1>
      <p><time datetime="%s">%s</time> %s</p>
      %s
    </header>

    %s

    <footer>shared via <a href="%s">gh-share</a></footer>
  </main>
</body>
</html>
`, themeAttr, escapedTitle, escapedDesc, escapedTitle, escapedDesc, escapedTitle, today, today, tagsHTML, ledeHTML, bodyHTML, basePagesURL)

	return doc, nil
}

// ListShares scans s/ directory for all shares
func ListShares(localDir string, basePagesURL string) ([]ShareItem, error) {
	sDir := filepath.Join(localDir, "s")
	entries, err := os.ReadDir(sDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var items []ShareItem
	titleReg := regexp.MustCompile(`(?i)<title>(.*?)</title>`)
	ledeReg := regexp.MustCompile(`(?i)<p class="lede">(.*?)</p>`)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		indexPath := filepath.Join(sDir, slug, "index.html")
		info, err := os.Stat(indexPath)
		if err != nil {
			continue
		}

		item := ShareItem{
			Slug:  slug,
			Path:  indexPath,
			MTime: info.ModTime(),
			URL:   fmt.Sprintf("%ss/%s/", strings.TrimRight(basePagesURL, "/")+"/", slug),
		}

		// Extract date from slug (YYYY-MM-DD)
		if len(slug) >= 10 && slug[4] == '-' && slug[7] == '-' {
			item.Date = slug[:10]
		} else {
			item.Date = info.ModTime().Format("2006-01-02")
		}

		// Parse title from index.html
		htmlBytes, err := os.ReadFile(indexPath)
		if err == nil {
			matches := titleReg.FindSubmatch(htmlBytes)
			if len(matches) > 1 {
				item.Title = html.UnescapeString(string(matches[1]))
			}
			ledeMatches := ledeReg.FindSubmatch(htmlBytes)
			if len(ledeMatches) > 1 {
				item.Lede = html.UnescapeString(string(ledeMatches[1]))
			}
		}

		if item.Title == "" {
			item.Title = slug
		}

		items = append(items, item)
	}

	// Sort newest first
	sort.Slice(items, func(i, j int) bool {
		if items[i].Date != items[j].Date {
			return items[i].Date > items[j].Date
		}
		return items[i].Slug > items[j].Slug
	})

	return items, nil
}
