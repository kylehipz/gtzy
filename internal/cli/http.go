package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func httpGet(path string) ([]byte, error) {
	resp, err := http.Get(BaseURL() + path)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s (is `gtzy serve` running?): %w", BaseURL(), err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, apiError(resp.StatusCode, body)
	}
	return body, nil
}

func httpPost(path string, payload any) ([]byte, error) {
	var buf bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&buf).Encode(payload); err != nil {
			return nil, err
		}
	}
	resp, err := http.Post(BaseURL()+path, "application/json", &buf)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s (is `gtzy serve` running?): %w", BaseURL(), err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, apiError(resp.StatusCode, body)
	}
	return body, nil
}

func apiError(status int, body []byte) error {
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &e) == nil && e.Error != "" {
		return fmt.Errorf("server: %s", e.Error)
	}
	return fmt.Errorf("server returned status %d", status)
}
