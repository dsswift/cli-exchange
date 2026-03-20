package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://graph.microsoft.com/v1.0"

type GraphClient struct {
	baseURL    string
	httpClient *http.Client
	tokenFn    func() (string, error)
}

func NewClient(baseURL string, timeout time.Duration, tokenFn func() (string, error)) *GraphClient {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &GraphClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
		tokenFn:    tokenFn,
	}
}

type graphErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *GraphClient) do(method, path string, params url.Values, body any) ([]byte, error) {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	token, err := c.tokenFn()
	if err != nil {
		return nil, &GraphAuthError{Message: err.Error()}
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var graphErr graphErrorResponse
		msg := string(respBody)
		if json.Unmarshal(respBody, &graphErr) == nil && graphErr.Error.Message != "" {
			msg = graphErr.Error.Message
		}

		switch resp.StatusCode {
		case 401, 403:
			return nil, &GraphAuthError{Message: msg}
		case 404:
			return nil, &GraphNotFoundError{Message: msg}
		default:
			return nil, &GraphError{StatusCode: resp.StatusCode, Message: msg}
		}
	}

	return respBody, nil
}

// doURL sends a request to an absolute URL (for following @odata.nextLink pagination).
func (c *GraphClient) doURL(method, fullURL string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	token, err := c.tokenFn()
	if err != nil {
		return nil, &GraphAuthError{Message: err.Error()}
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var graphErr graphErrorResponse
		msg := string(respBody)
		if json.Unmarshal(respBody, &graphErr) == nil && graphErr.Error.Message != "" {
			msg = graphErr.Error.Message
		}

		switch resp.StatusCode {
		case 401, 403:
			return nil, &GraphAuthError{Message: msg}
		case 404:
			return nil, &GraphNotFoundError{Message: msg}
		default:
			return nil, &GraphError{StatusCode: resp.StatusCode, Message: msg}
		}
	}

	return respBody, nil
}
