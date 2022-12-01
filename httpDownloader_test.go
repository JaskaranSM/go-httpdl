package httpdl

import (
	"net/http"
	"testing"
)

func TestResponseStatus(t *testing.T) {
	downloader := NewHTTPDownloader(&http.Client{})
	_, err := downloader.AddDownload("https://httpbin.org/status/400", &AddDownloadOpts{
		Connections: 10,
		Dir:         t.TempDir(),
	})
	if err == nil {
		t.Fatalf("Test failed, expected=err got nil")
	} else {
		t.Logf("Result: %v", err)
	}
}
