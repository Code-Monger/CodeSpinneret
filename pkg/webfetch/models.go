package webfetch

// WebFetchRequest represents a request to fetch a web page
type WebFetchRequest struct {
	// URL is the URL to fetch
	URL string `json:"url"`

	// IncludeImages determines whether to include images in the response
	IncludeImages bool `json:"include_images"`

	// StripHTML determines whether to strip all HTML tags from the content
	StripHTML bool `json:"strip_html"`

	// Timeout is the timeout in seconds
	Timeout int `json:"timeout"`
}

// WebFetchResponse represents the response from a web fetch operation
type WebFetchResponse struct {
	// URL is the URL that was fetched
	URL string `json:"url"`

	// StatusCode is the HTTP status code
	StatusCode int `json:"status_code"`

	// ContentType is the content type of the response
	ContentType string `json:"content_type"`

	// Headers contains the HTTP headers
	Headers map[string]string `json:"headers"`

	// Content is the content of the web page
	Content string `json:"content"`

	// Error is any error that occurred during the fetch
	Error string `json:"error,omitempty"`
}
