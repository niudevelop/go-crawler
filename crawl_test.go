package main

import (
	"net/url"
	"reflect"
	"testing"
)

func TestGetH1FromHTMLBasic(t *testing.T) {
	inputBody := "<html><body><h1>Test Title</h1></body></html>"
	actual := getH1FromHTML(inputBody)
	expected := "Test Title"

	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
func TestGetH1FromHTMLNoH1(t *testing.T) {
	inputBody := "<html><body><p>No header</p></body></html>"
	actual := getH1FromHTML(inputBody)
	expected := ""

	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestGetH1FromHTMLMultipleH1(t *testing.T) {
	inputBody := "<html><body><h1>First</h1><h1>Second</h1></body></html>"
	actual := getH1FromHTML(inputBody)
	expected := "First"

	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
func TestGetFirstParagraphFromHTMLMainPriority(t *testing.T) {
	inputBody := `<html><body>
		<p>Outside paragraph.</p>
		<main>
			<p>Main paragraph.</p>
		</main>
	</body></html>`
	actual := getFirstParagraphFromHTML(inputBody)
	expected := "Main paragraph."

	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestGetFirstParagraphFromHTMLNoMain(t *testing.T) {
	inputBody := `<html><body>
		<p>First paragraph.</p>
		<p>Second paragraph.</p>
	</body></html>`
	actual := getFirstParagraphFromHTML(inputBody)
	expected := "First paragraph."

	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestGetFirstParagraphFromHTMLNoParagraphs(t *testing.T) {
	inputBody := `<html><body>
		<main>
			<div>No paragraphs here</div>
		</main>
	</body></html>`
	actual := getFirstParagraphFromHTML(inputBody)
	expected := ""

	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestGetURLsFromHTMLNoLinks(t *testing.T) {
	base, _ := url.Parse("https://example.com")
	html := `<html><body><p>no anchors</p></body></html>`

	got, err := getURLsFromHTML(html, base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestGetURLsFromHTMLSkipsAnchorsWithoutHref(t *testing.T) {
	base, _ := url.Parse("https://example.com")
	html := `<html><body>
		<a>no href</a>
		<a name="x"></a>
	</body></html>`

	got, err := getURLsFromHTML(html, base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestGetURLsFromHTMLAbsoluteAndRelative(t *testing.T) {
	base, _ := url.Parse("https://example.com/base")
	html := `<html><body>
		<a href="https://other.com/x">abs</a>
		<a href="/about">rel</a>
	</body></html>`

	got, err := getURLsFromHTML(html, base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"https://other.com/x",
		"https://example.com/about",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d urls, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestGetURLsFromHTMLSkipsInvalidHref(t *testing.T) {
	base, _ := url.Parse("https://example.com")
	html := `<html><body>
		<a href="https://ok.com/a">ok</a>
		<a href="http://[::1">bad</a>
		<a href="/b">rel</a>
	</body></html>`

	got, err := getURLsFromHTML(html, base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"https://ok.com/a",
		"https://example.com/b",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d urls, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestGetImagesFromHTMLRelative(t *testing.T) {
	inputURL := "https://blog.boot.dev"
	inputBody := `<html><body><img src="/logo.png" alt="Logo"></body></html>`

	baseURL, err := url.Parse(inputURL)
	if err != nil {
		t.Errorf("couldn't parse input URL: %v", err)
		return
	}

	actual, err := getImagesFromHTML(inputBody, baseURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"https://blog.boot.dev/logo.png"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestExtractPageData(t *testing.T) {
	tests := []struct {
		name     string
		pageURL  string
		html     string
		expected PageData
	}{
		{
			name:    "happy path: h1, first p, one relative link, one relative image",
			pageURL: "https://blog.boot.dev",
			html: `<html><body>
				<h1>Test Title</h1>
				<p>This is the first paragraph.</p>
				<a href="/link1">Link 1</a>
				<img src="/image1.jpg" alt="Image 1">
			</body></html>`,
			expected: PageData{
				URL:            "https://blog.boot.dev",
				H1:             "Test Title",
				FirstParagraph: "This is the first paragraph.",
				OutgoingLinks:  []string{"https://blog.boot.dev/link1"},
				ImageURLs:      []string{"https://blog.boot.dev/image1.jpg"},
			},
		},
		{
			name:    "absolute link and absolute image preserved",
			pageURL: "https://blog.boot.dev",
			html: `<html><body>
				<h1>Title</h1>
				<p>Para</p>
				<a href="https://example.com/a">A</a>
				<img src="https://cdn.example.com/i.png">
			</body></html>`,
			expected: PageData{
				URL:            "https://blog.boot.dev",
				H1:             "Title",
				FirstParagraph: "Para",
				OutgoingLinks:  []string{"https://example.com/a"},
				ImageURLs:      []string{"https://cdn.example.com/i.png"},
			},
		},
		{
			name:    "base URL with path: relative links resolve against origin (not the path)",
			pageURL: "https://blog.boot.dev/some/path",
			html: `<html><body>
				<h1>Title</h1>
				<p>Para</p>
				<a href="link1">Link 1</a>
				<a href="/link2">Link 2</a>
				<img src="image.jpg">
				<img src="/img2.jpg">
			</body></html>`,
			expected: PageData{
				URL:            "https://blog.boot.dev/some/path",
				H1:             "Title",
				FirstParagraph: "Para",
				OutgoingLinks:  []string{"https://blog.boot.dev/link1", "https://blog.boot.dev/link2"},
				ImageURLs:      []string{"https://blog.boot.dev/image.jpg", "https://blog.boot.dev/img2.jpg"},
			},
		},
		{
			name:    "missing elements: empty h1 and paragraph, no links/images",
			pageURL: "https://blog.boot.dev",
			html:    `<html><body><div>No relevant tags</div></body></html>`,
			expected: PageData{
				URL:            "https://blog.boot.dev",
				H1:             "",
				FirstParagraph: "",
				OutgoingLinks:  []string{},
				ImageURLs:      []string{},
			},
		},
		{
			name:    "multiple h1 and p: first ones used; multiple links/images collected",
			pageURL: "https://blog.boot.dev",
			html: `<html><body>
				<h1>First H1</h1>
				<h1>Second H1</h1>
				<p>First paragraph.</p>
				<p>Second paragraph.</p>
				<a href="/a">A</a>
				<a href="/b">B</a>
				<img src="/1.png">
				<img src="/2.png">
			</body></html>`,
			expected: PageData{
				URL:            "https://blog.boot.dev",
				H1:             "First H1",
				FirstParagraph: "First paragraph.",
				OutgoingLinks:  []string{"https://blog.boot.dev/a", "https://blog.boot.dev/b"},
				ImageURLs:      []string{"https://blog.boot.dev/1.png", "https://blog.boot.dev/2.png"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := extractPageData(tc.html, tc.pageURL)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, actual)
			}
		})
	}
}
