package websearch

// SearchResults represents the results of a web search
type SearchResults struct {
	Query   string
	Engine  string
	Results []SearchResult
}

// SearchResult represents a single search result
type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

// DuckDuckGoResponse represents the response from DuckDuckGo API
type DuckDuckGoResponse struct {
	AbstractText  string   `json:"AbstractText"`
	AbstractURL   string   `json:"AbstractURL"`
	Definition    string   `json:"Definition"`
	DefinitionURL string   `json:"DefinitionURL"`
	Heading       string   `json:"Heading"`
	RelatedTopics []Topic  `json:"RelatedTopics"`
	Results       []Result `json:"Results"`
	Type          string   `json:"Type"`
}

// Topic represents a related topic in DuckDuckGo response
type Topic struct {
	FirstURL string  `json:"FirstURL"`
	Icon     Icon    `json:"Icon"`
	Result   string  `json:"Result"`
	Text     string  `json:"Text"`
	Topics   []Topic `json:"Topics,omitempty"`
}

// Icon represents an icon in DuckDuckGo response
type Icon struct {
	Height string `json:"Height"`
	URL    string `json:"URL"`
	Width  string `json:"Width"`
}

// Result represents a result in DuckDuckGo response
type Result struct {
	FirstURL string `json:"FirstURL"`
	Icon     Icon   `json:"Icon"`
	Result   string `json:"Result"`
	Text     string `json:"Text"`
}

// BingResponse represents the response from Bing API
type BingResponse struct {
	WebPages struct {
		Value []BingWebPage `json:"value"`
	} `json:"webPages"`
}

// BingWebPage represents a web page in Bing response
type BingWebPage struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	DisplayURL      string `json:"displayUrl"`
	Snippet         string `json:"snippet"`
	DateLastCrawled string `json:"dateLastCrawled"`
}

// GoogleResponse represents the response from Google API
type GoogleResponse struct {
	Items []GoogleItem `json:"items"`
}

// GoogleItem represents an item in Google response
type GoogleItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	DisplayLink string `json:"displayLink"`
	Snippet     string `json:"snippet"`
}
