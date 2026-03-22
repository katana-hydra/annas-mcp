package mirror

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

// MirrorStatusURL is the source for dynamic mirror list
const MirrorStatusURL = "https://open-slum.org/"

// Hardcoded fallback mirrors (used if dynamic fetch fails)
var fallbackMirrors = []string{
	"annas-archive.gl",
	"annas-archive.pk",
	"annas-archive.gd",
	"annas-archive.vg",
	"annas-archive.se",
	"annas-archive.li",
}

// Mirror represents a single mirror with its health status
type Mirror struct {
	Name         string
	IsWorking    bool
	ResponseTime time.Duration
}

// Selector manages mirror selection with automatic fallback
type Selector struct {
	mirrors   []string
	results   []Mirror
	mu        sync.RWMutex
	lastCheck time.Time
}

// NewSelector creates a new mirror selector and performs initial health check
func NewSelector() *Selector {
	s := &Selector{}
	
	// Try to fetch mirrors dynamically from open-slum.org
	dynamicMirrors := fetchMirrorsFromStatusPage()
	if len(dynamicMirrors) > 0 {
		s.mirrors = dynamicMirrors
	} else {
		s.mirrors = fallbackMirrors
	}
	
	s.refreshMirrors()
	return s
}

// fetchMirrorsFromStatusPage scrapes mirror list from open-slum.org
func fetchMirrorsFromStatusPage() []string {
	var mirrors []string
	var mu sync.Mutex
	
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)
	
	// Pattern to match annas-archive domains (supports .gl, .pk, .gd, .vg, .se, .li, etc.)
	domainRegex := regexp.MustCompile(`annas-archive\.[a-z]+`)
	
	// Helper to add mirror if not already present
	addMirror := func(domain string) {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			return
		}
		
		mu.Lock()
		defer mu.Unlock()
		
		for _, m := range mirrors {
			if m == domain {
				return
			}
		}
		mirrors = append(mirrors, domain)
	}
	
	// Extract domain from various sources
	extractDomain := func(text string) {
		// Remove protocol if present
		text = strings.TrimPrefix(text, "https://")
		text = strings.TrimPrefix(text, "http://")
		
		// Find domain pattern
		if domainRegex.MatchString(text) {
			domain := domainRegex.FindString(text)
			if domain != "" {
				addMirror(domain)
			}
		}
	}
	
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		extractDomain(e.Attr("href"))
	})
	
	c.OnHTML("td, div, span, p, li", func(e *colly.HTMLElement) {
		extractDomain(e.Text)
	})
	
	// Set timeout
	c.SetRequestTimeout(15 * time.Second)
	
	// Visit the status page
	err := c.Visit(MirrorStatusURL)
	if err != nil {
		// Return empty to trigger fallback
		return nil
	}
	
	c.Wait()
	
	return mirrors
}

// refreshMirrors tests all mirrors and updates the working list
func (s *Selector) refreshMirrors() {
	results := s.testAllMirrors()

	// Sort by response time (working mirrors first, then by speed)
	sort.Slice(results, func(i, j int) bool {
		if results[i].IsWorking != results[j].IsWorking {
			return results[i].IsWorking && !results[j].IsWorking
		}
		return results[i].ResponseTime < results[j].ResponseTime
	})

	s.mu.Lock()
	s.results = results
	s.lastCheck = time.Now()
	s.mu.Unlock()
}

// testAllMirrors concurrently tests all mirrors
func (s *Selector) testAllMirrors() []Mirror {
	var wg sync.WaitGroup
	resultChan := make(chan Mirror, len(s.mirrors))

	for _, mirror := range s.mirrors {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			result := s.testMirror(m)
			resultChan <- result
		}(mirror)
	}

	wg.Wait()
	close(resultChan)

	var results []Mirror
	for r := range resultChan {
		results = append(results, r)
	}

	return results
}

// testMirror tests a single mirror's availability
func (s *Selector) testMirror(mirror string) Mirror {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	start := time.Now()
	
	url := "https://" + mirror
	resp, err := client.Head(url)
	if err != nil {
		return Mirror{Name: mirror, IsWorking: false, ResponseTime: time.Since(start)}
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx status codes as working
	working := resp.StatusCode >= 200 && resp.StatusCode < 400

	return Mirror{
		Name:         mirror,
		IsWorking:    working,
		ResponseTime: time.Since(start),
	}
}

// GetBestMirror returns the best available mirror
func (s *Selector) GetBestMirror() (*Mirror, error) {
	s.mu.RLock()
	results := s.results
	lastCheck := s.lastCheck
	s.mu.RUnlock()

	// Refresh if no results or cache is older than 5 minutes
	if len(results) == 0 || time.Since(lastCheck) > 5*time.Minute {
		s.refreshMirrors()
		s.mu.RLock()
		results = s.results
		s.mu.RUnlock()
	}

	for _, r := range results {
		if r.IsWorking {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("no working mirrors available")
}

// GetFallbackChain returns all mirrors sorted by priority (working first)
func (s *Selector) GetFallbackChain() []Mirror {
	s.mu.RLock()
	results := s.results
	lastCheck := s.lastCheck
	s.mu.RUnlock()

	// Refresh if cache is older than 5 minutes
	if time.Since(lastCheck) > 5*time.Minute {
		s.refreshMirrors()
		s.mu.RLock()
		results = s.results
		s.mu.RUnlock()
	}

	return results
}

// ForceRefresh forces an immediate mirror health check
func (s *Selector) ForceRefresh() {
	// Re-fetch mirrors from status page
	dynamicMirrors := fetchMirrorsFromStatusPage()
	if len(dynamicMirrors) > 0 {
		s.mirrors = dynamicMirrors
	}
	s.refreshMirrors()
}

// GetMirrorList returns the current list of mirrors (for debugging)
func (s *Selector) GetMirrorList() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mirrors
}

// ExtractBaseURL extracts the base URL (domain) from a full URL
func ExtractBaseURL(fullURL string) string {
	// Remove protocol
	fullURL = strings.TrimPrefix(fullURL, "https://")
	fullURL = strings.TrimPrefix(fullURL, "http://")
	// Remove path
	if idx := strings.Index(fullURL, "/"); idx > 0 {
		fullURL = fullURL[:idx]
	}
	return fullURL
}
