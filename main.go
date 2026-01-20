package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

type config struct {
	pages              map[string]PageData
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 1 {
		log.Fatal("no website provided")
	}
	// if len(argsWithoutProg) > 1 {
	// 	log.Fatal("too many arguments provided")
	// }
	baseURL := argsWithoutProg[0]
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	maxConcurrency, err := strconv.Atoi(argsWithoutProg[1])
	if err != nil || maxConcurrency <= 0 {
		log.Fatal("maxConcurrency must be a positive integer")
	}

	maxPages, err := strconv.Atoi(argsWithoutProg[2])
	if err != nil || maxPages <= 0 {
		log.Fatal("maxPages must be a positive integer")
	}

	fmt.Printf("starting crawl of: %s\n", baseURL)
	cfg := config{
		pages:              make(map[string]PageData),
		baseURL:            parsedBaseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}
	// pages := make(map[string]int)
	cfg.wg.Add(1)
	go cfg.crawlPage(baseURL)

	cfg.wg.Wait()
	if err := writeCSVReport(cfg.pages, "report.csv"); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(pages)
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

// func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int) {
func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.mu.Lock()
	if len(cfg.pages) >= cfg.maxPages {
		cfg.mu.Unlock()
		// <-cfg.concurrencyControl
		cfg.wg.Done()
		return
	}
	cfg.mu.Unlock()

	cfg.concurrencyControl <- struct{}{}
	defer func() {
		<-cfg.concurrencyControl
		cfg.wg.Done()
	}()
	parseCurrentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	if cfg.baseURL.Host != parseCurrentURL.Host {
		return
	}
	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		return
	}
	isFirst := cfg.addPageVisit(normalizedURL)
	if !isFirst {
		return
	}

	fmt.Printf("Crawling: %s\n", rawCurrentURL)
	html, err := getHTML(rawCurrentURL)
	if err != nil {
		return
	}
	data := extractPageData(html, rawCurrentURL)
	cfg.setPageData(normalizedURL, data)
	// outgoingLinks, err := getURLsFromHTML(html, cfg.baseURL)
	// if err != nil {
	// 	return
	// }
	for _, v := range cfg.pages[normalizedURL].OutgoingLinks {
		cfg.wg.Add(1)
		go cfg.crawlPage(v)
	}
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if _, visited := cfg.pages[normalizedURL]; visited {
		return false
	}

	cfg.pages[normalizedURL] = PageData{URL: normalizedURL}
	return true
}

// setPageData safely stores the final PageData for a URL.
func (cfg *config) setPageData(normalizedURL string, data PageData) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.pages[normalizedURL] = data
}

func configure(rawBaseURL string, maxConcurrency int) (*config, error) {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse base URL: %v", err)
	}

	return &config{
		pages:              make(map[string]PageData),
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
	}, nil
}
