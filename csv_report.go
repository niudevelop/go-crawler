package main

import (
	"encoding/csv"
	"os"
	"sort"
	"strings"
)

func writeCSVReport(pages map[string]PageData, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{
		"page_url",
		"h1",
		"first_paragraph",
		"outgoing_link_urls",
		"image_urls",
	}); err != nil {
		return err
	}

	urls := make([]string, 0, len(pages))
	for u := range pages {
		urls = append(urls, u)
	}
	sort.Strings(urls)

	for _, u := range urls {
		p := pages[u]

		outgoing := strings.Join(p.OutgoingLinks, ";")
		images := strings.Join(p.ImageURLs, ";")

		if err := w.Write([]string{
			p.URL,
			p.H1,
			p.FirstParagraph,
			outgoing,
			images,
		}); err != nil {
			return err
		}
	}

	return w.Error()
}
