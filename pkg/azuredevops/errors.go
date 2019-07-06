package azuredevops

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPError is returned when an HTTP response does not return 200
type HTTPError struct {
	StatusCode int

	Endpoint string

	RetryAfter *time.Duration
}

// NewHTTPError returns an HTTPError
func NewHTTPError(response *http.Response) *HTTPError {
	var retryAfter *time.Duration
	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode == http.StatusServiceUnavailable {
		retryAfterStr, exists := response.Header["Retry-After"]
		if exists {
			retryAfterVal, err := time.ParseDuration(fmt.Sprintf("%ss", retryAfterStr))
			if err != nil {
				retryAfter = &retryAfterVal
			}
		}
	}

	return &HTTPError{
		StatusCode: response.StatusCode,
		Endpoint:   response.Request.URL.Path,
		RetryAfter: retryAfter,
	}
}

func (err HTTPError) Error() string {
	return fmt.Sprintf("Error - received HTTP status code %d when calling call to %s", err.StatusCode, err.Endpoint)
}
