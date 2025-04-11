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

// searchDuckDuckGo performs a search using DuckDuckGo
func searchDuckDuckGo(query string, numResults int, safeSearch bool) (*SearchResults, error) {
	config := GetConfig()

	// Build the request URL like a human would
	baseURL := config.DuckDuckGoURL
	params := url.Values{}
	params.Add("q", query)

	if safeSearch {
		params.Add("kp", "1") // Safe search enabled
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
		return nil, fmt.Errorf("error response from DuckDuckGo: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Parse the HTML response
	htmlContent := string(body)

	// Save the HTML content to a file for debugging
	debugFile := fmt.Sprintf("duckduckgo_response_%s.html", strings.ReplaceAll(query, " ", "_"))
	err = os.WriteFile(debugFile, []byte(htmlContent), 0644)
	if err != nil {
		log.Printf("Warning: Could not save debug file: %v", err)
	} else {
		log.Printf("Saved DuckDuckGo response to %s", debugFile)
	}

	// Create results structure
	results := &SearchResults{
		Query:   query,
		Engine:  "DuckDuckGo",
		Results: []SearchResult{},
	}

	// Look for result snippets
	// Based on the HTML example provided by the user
	snippetRegex := regexp.MustCompile(`<a class="result__snippet" href="([^"]+)">(.*?)</a>`)
	snippetMatches := snippetRegex.FindAllStringSubmatch(htmlContent, -1)

	log.Printf("Found %d snippet matches", len(snippetMatches))

	// Process snippet matches
	for _, match := range snippetMatches {
		if len(match) >= 3 {
			url := match[1]
			snippet := cleanHTML(match[2])

			// Try to find the title for this result
			titleRegex := regexp.MustCompile(fmt.Sprintf(`<a[^>]*href="%s"[^>]*>(.*?)</a>.*?<a class="result__snippet"`, regexp.QuoteMeta(url)))
			titleMatch := titleRegex.FindStringSubmatch(htmlContent)

			title := ""
			if len(titleMatch) >= 2 {
				title = cleanHTML(titleMatch[1])
			}

			// If we couldn't find a title, try to find the display URL
			if title == "" {
				urlRegex := regexp.MustCompile(fmt.Sprintf(`<a class="result__url" href="%s">(.*?)</a>`, regexp.QuoteMeta(url)))
				urlMatch := urlRegex.FindStringSubmatch(htmlContent)

				if len(urlMatch) >= 2 {
					title = strings.TrimSpace(urlMatch[1])
				} else {
					// If we still couldn't find a title, use the URL
					title = url
				}
			}

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

	// If we couldn't find results with the first approach, try an alternative approach
	if len(results.Results) == 0 {
		log.Printf("No results found with first approach, trying alternative...")

		// Look for result divs
		resultRegex := regexp.MustCompile(`<div class="result[^"]*">.*?<a class="result__a" href="([^"]+)"[^>]*>(.*?)</a>.*?<a class="result__snippet"[^>]*>(.*?)</a>`)
		resultMatches := resultRegex.FindAllStringSubmatch(htmlContent, -1)

		log.Printf("Found %d result matches with alternative approach", len(resultMatches))

		for _, match := range resultMatches {
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
				log.Printf("Found result (alt): Title=%s, URL=%s, Snippet=%s", title, url, snippet)

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

		// Look for any links as a last resort
		linkRegex := regexp.MustCompile(`<a\s+[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
		linkMatches := linkRegex.FindAllStringSubmatch(htmlContent, -1)

		for _, match := range linkMatches {
			if len(match) >= 3 && !strings.Contains(match[1], "javascript:") && !strings.HasPrefix(match[1], "#") {
				url := match[1]
				title := cleanHTML(match[2])

				// Skip empty titles or navigation links
				if title == "" || len(title) < 5 || strings.Contains(strings.ToLower(title), "next") || strings.Contains(strings.ToLower(title), "previous") {
					continue
				}

				// Add to results
				results.Results = append(results.Results, SearchResult{
					Title:   title,
					URL:     url,
					Snippet: "No snippet available",
				})

				// Log the found result
				log.Printf("Found result (fallback): Title=%s, URL=%s", title, url)

				// Check if we have enough results
				if len(results.Results) >= numResults {
					break
				}
			}
		}
	}

	return results, nil
}
