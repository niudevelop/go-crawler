package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type config struct {
	pages              map[string]PageData
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
}

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 1 {
		log.Fatal("no website provided")
	}
	if len(argsWithoutProg) > 1 {
		log.Fatal("too many arguments provided")
	}
	baseURL := argsWithoutProg[0]
	fmt.Printf("starting crawl of: %s\n", baseURL)
	pages := make(map[string]int)
	crawlPage(baseURL, baseURL, pages)
	fmt.Println(pages)
	// html, err := getHTML(baseURL)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	// fmt.Println(html)

}

func getHTML(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "BootCrawler/1.0")

	client := http.DefaultClient

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode > 400 {
		return "", fmt.Errorf("%s", res.Status)
	}
	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return "", fmt.Errorf("Unsupported content type: %s", contentType)
	}
	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(resData), nil

}

func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int) {
	parseBaseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	parseCurrentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	if parseBaseURL.Host != parseCurrentURL.Host {
		return
	}
	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		return
	}
	_, ok := pages[normalizedURL]
	if !ok {
		pages[normalizedURL] = 1
	} else {
		pages[normalizedURL] += 1
		return
	}
	fmt.Printf("Crawling: %s\n", rawCurrentURL)
	html, err := getHTML(rawCurrentURL)
	if err != nil {
		return
	}
	outgoingLinks, err := getURLsFromHTML(html, parseBaseURL)
	if err != nil {
		return
	}
	for _, v := range outgoingLinks {
		if v != rawCurrentURL {
			crawlPage(rawBaseURL, v, pages)
		}
	}
}
