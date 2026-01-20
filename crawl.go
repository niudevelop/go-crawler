package main

import (
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PageData struct {
	URL            string
	H1             string
	FirstParagraph string
	OutgoingLinks  []string
	ImageURLs      []string
}

func getH1FromHTML(html string) string {
	doc := getQueryDoc(html)
	node := doc.Find("h1")
	return node.First().Text()
}

func getFirstParagraphFromHTML(html string) string {
	doc := getQueryDoc(html)
	mainNodes := doc.Find("main p")
	if mainNodes.Length() == 0 {
		pNodes := doc.Find("p")
		return pNodes.First().Text()
	}
	return mainNodes.First().Text()
}

func getQueryDoc(html string) *goquery.Document {
	body := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func getURLsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	doc := getQueryDoc(htmlBody)
	aNodes := doc.Find("a")
	if aNodes.Length() == 0 {
		return []string{}, nil
	}
	var urls []string
	aNodes.Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		parsed, err := url.Parse(href)
		if err != nil {
			return
		}
		isRelative := parsed.Host == "" && parsed.Scheme == ""
		if isRelative {
			baseURL.Path = parsed.Path
			urls = append(urls, baseURL.String())
		} else {
			urls = append(urls, parsed.String())
		}
	})
	return urls, nil
}

func getImagesFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	doc := getQueryDoc(htmlBody)
	aNodes := doc.Find("img")
	if aNodes.Length() == 0 {
		return []string{}, nil
	}
	var urls []string
	aNodes.Each(func(i int, s *goquery.Selection) {
		src, ok := s.Attr("src")
		if !ok {
			return
		}
		parsed, err := url.Parse(src)
		if err != nil {
			return
		}
		isRelative := parsed.Host == "" && parsed.Scheme == ""
		if isRelative {
			baseURL.Path = parsed.Path
			urls = append(urls, baseURL.String())
		} else {
			urls = append(urls, parsed.String())
		}
	})
	return urls, nil
}

func extractPageData(html, pageURL string) PageData {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	outgoingLinks, _ := getURLsFromHTML(html, parsedURL)
	imageURLs, _ := getImagesFromHTML(html, parsedURL)
	return PageData{
		URL:            pageURL,
		H1:             getH1FromHTML(html),
		FirstParagraph: getFirstParagraphFromHTML(html),
		OutgoingLinks:  outgoingLinks,
		ImageURLs:      imageURLs,
	}
}
