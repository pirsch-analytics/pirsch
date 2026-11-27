package ip

import (
	"fmt"
	"io"
	"net/http"
)

func checkDownloadResponse(resp *http.Response) error {
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("error downloading database: unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
