package graph

import "fmt"

type GraphError struct {
	StatusCode int
	Message    string
}

func (e *GraphError) Error() string {
	return fmt.Sprintf("graph API error (%d): %s", e.StatusCode, e.Message)
}

type GraphNotFoundError struct {
	Message string
}

func (e *GraphNotFoundError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("not found: %s", e.Message)
	}
	return "resource not found"
}

type GraphAuthError struct {
	Message string
}

func (e *GraphAuthError) Error() string {
	return fmt.Sprintf("authentication error: %s", e.Message)
}
