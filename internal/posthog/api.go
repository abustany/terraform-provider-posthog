// Package posthog implements a client for the PostHog API.
package posthog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	HTTPClient HTTPClient
	Host       string
	APIKey     string
}

type apiRequest struct {
	Method         string
	Path           string
	ExpectedCode   int
	Input          any
	Output         any
	OutputNilIf404 bool
}

func (c *Client) do(ctx context.Context, r apiRequest) error {
	var body io.Reader

	if r.Input != nil {
		j, err := json.Marshal(r.Input)
		if err != nil {
			return fmt.Errorf("error marshalling input to json: %w", err)
		}
		body = bytes.NewReader(j)
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, c.Host+"/api"+r.Path, body)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	if r.Output != nil {
		req.Header.Add("Accept", "application/json")
	}

	req.Header.Add("Authorization", "Bearer "+c.APIKey)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error doing HTTP request: %w", err)
	}

	defer res.Body.Close()

	var responseBody io.Reader

	if r.OutputNilIf404 && res.StatusCode == 404 {
		// nothing to unmarshal, output will be nil
		responseBody = strings.NewReader("null") // ¯\_(ツ)_/¯
	} else if res.StatusCode != r.ExpectedCode {
		errorBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected HTTP status code, expected %d, got %d (server response: %s)", r.ExpectedCode, res.StatusCode, string(errorBody))
	} else { // res.StatusCode == r.ExpectedCode
		responseBody = res.Body
	}

	if r.Output != nil {
		if err := json.NewDecoder(responseBody).Decode(r.Output); err != nil {
			return fmt.Errorf("error decoding JSON reply: %w", err)
		}
	}

	return nil
}
