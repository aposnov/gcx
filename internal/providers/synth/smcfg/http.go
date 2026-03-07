package smcfg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleErrorResponse reads a non-2xx HTTP response body and returns a descriptive error.
// It attempts to decode the SM API's JSON error format ({error, msg}) before falling
// back to the raw body or a status-code-only message.
func HandleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("request failed with status %d (could not read body: %w)", resp.StatusCode, err)
	}

	var errResp struct {
		Error string `json:"error"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Error != "" {
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, errResp.Error)
		}
		if errResp.Msg != "" {
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, errResp.Msg)
		}
	}

	if len(body) > 0 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("request failed with status %d", resp.StatusCode)
}
