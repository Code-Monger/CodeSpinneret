package websearch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// searchBing performs a search using Bing
func searchBing(query string, numResults int, safeSearch bool) (*SearchResults, error) {
	config := GetConfig()

	// Build the request URL like a human would
	baseURL := config.BingURL
	params := url.Values{}
	params.Add("q", query)
	params.Add("count", fmt.Sprintf("%d", numResults))

	if safeSearch {
		params.Add("safesearch", "strict")
	} else {
		params.Add("safesearch", "off")
	}

	// Create the request
	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers to mimic a browser
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from Bing: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Parse the HTML response
	htmlContent := string(body)

	// Save the HTML content to a file for debugging
	debugFile := fmt.Sprintf("bing_response_%s.html", strings.ReplaceAll(query, " ", "_"))
	err = os.WriteFile(debugFile, []byte(htmlContent), 0644)
	if err != nil {
		log.Printf("Warning: Could not save debug file: %v", err)
	} else {
		log.Printf("Saved Bing response to %s", debugFile)
	}

	// Create results structure
	results := &SearchResults{
		Query:   query,
		Engine:  "Bing",
		Results: []SearchResult{},
	}

	// Based on the successful "artificial intelligence" query, we can see that
	// Bing search results are in <li class="b_algo"> elements
	// The title is in an <h2><a href="..."> element
	// The snippet is in a <p class="b_lineclamp2"> element

	// Extract search results based on the actual HTML structure
	resultRegex := regexp.MustCompile(`<li class="b_algo"[^>]*>.*?<h2><a href="([^"]+)"[^>]*>(.*?)</a></h2>.*?<p class="b_lineclamp2">(.*?)</p>`)
	matches := resultRegex.FindAllStringSubmatch(htmlContent, -1)

	log.Printf("Found %d search results with main pattern", len(matches))

	// Process matches
	for _, match := range matches {
		if len(match) >= 4 {
			url := match[1]
			title := cleanHTML(match[2])
			snippet := cleanHTML(match[3])

			// Add to results
			results.Results = append(results.Results, SearchResult{
				Title:   title,
				URL:     url,
				Snippet: snippet,
			})

			// Log the found result
			log.Printf("Found result: Title=%s, URL=%s, Snippet=%s", title, url, snippet)

			// Check if we have enough results
			if len(results.Results) >= numResults {
				break
			}
		}
	}

	// If we couldn't find results with the first pattern, try an alternative pattern
	// This pattern is more flexible and might work for different queries
	if len(results.Results) == 0 {
		log.Printf("No results found with main pattern, trying alternative pattern...")

		// Try to extract title, URL, and snippet separately
		// Look for h2 tags with links
		h2Regex := regexp.MustCompile(`<h2><a href="([^"]+)"[^>]*>(.*?)</a></h2>`)
		h2Matches := h2Regex.FindAllStringSubmatch(htmlContent, -1)

		log.Printf("Found %d h2 matches with alternative pattern", len(h2Matches))

		for _, match := range h2Matches {
			if len(match) >= 3 {
				url := match[1]
				title := cleanHTML(match[2])

				// Try to find a snippet for this URL
				// Look for a paragraph after the h2 tag
				snippetRegex := regexp.MustCompile(fmt.Sprintf(`<h2><a href="%s"[^>]*>.*?</a></h2>.*?<p[^>]*>(.*?)</p>`, regexp.QuoteMeta(url)))
				snippetMatches := snippetRegex.FindStringSubmatch(htmlContent)

				snippet := "No snippet available"
				if len(snippetMatches) >= 2 {
					snippet = cleanHTML(snippetMatches[1])
				}

				// Add to results
				results.Results = append(results.Results, SearchResult{
					Title:   title,
					URL:     url,
					Snippet: snippet,
				})

				// Log the found result
				log.Printf("Found result (alt): Title=%s, URL=%s, Snippet=%s", title, url, snippet)

				// Check if we have enough results
				if len(results.Results) >= numResults {
					break
				}
			}
		}
	}

	// If we still couldn't find any results, try a more general approach
	if len(results.Results) == 0 {
		log.Printf("No results found with alternative pattern, trying general approach...")

		// Look for any links with titles in h2 tags, more broadly
		h2Regex := regexp.MustCompile(`<h2[^>]*>.*?<a[^>]*href="([^"]+)"[^>]*>(.*?)</a>.*?</h2>`)
		h2Matches := h2Regex.FindAllStringSubmatch(htmlContent, -1)

		log.Printf("Found %d h2 matches with general approach", len(h2Matches))

		for _, match := range h2Matches {
			if len(match) >= 3 {
				url := match[1]
				title := cleanHTML(match[2])

				// Add to results
				results.Results = append(results.Results, SearchResult{
					Title:   title,
					URL:     url,
					Snippet: "No snippet available",
				})

				// Log the found result
				log.Printf("Found result (general): Title=%s, URL=%s", title, url)

				// Check if we have enough results
				if len(results.Results) >= numResults {
					break
				}
			}
		}
	}

	// If we still couldn't find any results, log the HTML structure
	if len(results.Results) == 0 {
		log.Printf("No results found. HTML structure might have changed. Check the debug file: %s", debugFile)

		// Extract the first 1000 characters of the HTML for logging
		htmlPreview := htmlContent
		if len(htmlPreview) > 1000 {
			htmlPreview = htmlPreview[:1000]
		}
		log.Printf("HTML preview: %s", htmlPreview)
	}

	return results, nil
}
