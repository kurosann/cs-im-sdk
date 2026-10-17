package csim

import "fmt"

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("cs-im api error: status=%d code=%s message=%s", e.StatusCode, e.Code, e.Message)
}
