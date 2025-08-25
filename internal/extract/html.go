package extract

// genspark: HTML fetching and extraction utilities

import (
	"context"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// genspark: Result holds extracted page data
type Result struct {
	URL      string   `json:"url"`
	Title    string   `json:"title"`
	Text     string   `json:"text"`
	Headings []string `json:"headings"`
	Links    []string `json:"links"`
	Prices   []string `json:"prices"`
}

// genspark: FetchAndExtract downloads and parses a web page
func FetchAndExtract(ctx context.Context, url string) (*Result, error) {
	client := &http.Client{Timeout: 12 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "genspark-mini/1.0") // genspark
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("non-2xx status: " + resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	root, err := html.Parse(strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	title := getTitle(root)
	text := getMainText(root)
	headings := getHeadings(root, map[string]bool{"h1": true, "h2": true, "h3": true})
	links := getLinks(root)
	prices := findPrices(text)
	return &Result{
		URL:      url,
		Title:    title,
		Text:     text,
		Headings: headings,
		Links:    links,
		Prices:   prices,
	}, nil
}

// genspark: extract <title>
func getTitle(n *html.Node) string {
	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = strings.TrimSpace(n.FirstChild.Data)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return title
}

// genspark: collect text content (very naive main-text heuristic)
func getMainText(n *html.Node) string {
	var parts []string
	var skip = map[string]bool{"script": true, "style": true, "nav": true, "footer": true, "header": true}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && skip[n.Data] {
			return
		}
		if n.Type == html.TextNode {
			val := strings.TrimSpace(n.Data)
			if val != "" {
				parts = append(parts, val)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	// genspark: join and collapse spaces
	joined := strings.Join(parts, " ")
	joined = regexp.MustCompile(`\s+`).ReplaceAllString(joined, " ")
	return strings.TrimSpace(joined)
}

// genspark: gather headings
func getHeadings(n *html.Node, allow map[string]bool) []string {
	var out []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && allow[n.Data] {
			if n.FirstChild != nil {
				out = append(out, strings.TrimSpace(innerText(n)))
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return out
}

// genspark: get all <a href>
func getLinks(n *html.Node) []string {
	var out []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					val := strings.TrimSpace(a.Val)
					if val != "" {
						out = append(out, val)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	// genspark: unique + stable sort
	m := map[string]struct{}{}
	var uniq []string
	for _, v := range out {
		if _, ok := tm[v]; !ok {
			tm[v] = struct{}{}; uniq = append(uniq, v)
		}
	}
	sort.Strings(uniq)
	return uniq
}

// genspark: innerText helper
func innerText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(innerText(c))
		if c.Type == html.ElementNode {
			b.WriteString(" ")
		}
	}
	return strings.TrimSpace(b.String())
}

// genspark: detect naive price patterns
func findPrices(text string) []string {
	re := regexp.MustCompile(`(?i)(?:\$|€|£)\s?\d+[\d,\.]*`)
	m := re.FindAllString(text, -1)
	// genspark: unique
	seen := map[string]struct{}{}
	var out []string
	for _, v := range m {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
