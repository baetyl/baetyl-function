package common

const (
	// HTTPVerb is an dapr http channel verb
	HTTPVerb = "http.verb"
	// HTTPStatusCode is an dapr http channel status code
	HTTPStatusCode = "http.status_code"
	// Get is an HTTP Get method
	Get = "GET"
	// Post is an Post Get method
	Post = "POST"
	// Delete is an HTTP Delete method
	Delete = "DELETE"
	// Put is an HTTP Put method
	Put = "PUT"
	// Options is an HTTP OPTIONS method
	Options = "OPTIONS"
	// QueryString is the query string passed by the request
	QueryString = "http.query_string"
	// ContentType is the header for Content-Type
	ContentType        = "Content-Type"
	defaultContentType = "application/json"

	httpInternalErrorCode = "500"
	HeaderDelim           = "&__header_delim__&"
	HeaderEquals          = "&__header_equals__&"
)
