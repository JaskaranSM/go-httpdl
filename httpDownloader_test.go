package httpdl

import (
	"net/http"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
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

type TestHTTPDownloadListener struct {
	startTime              time.Time
	progressSpinnerRunning bool
	t                      *testing.T
	exitChan               chan struct{}
}

func (t *TestHTTPDownloadListener) progressSpinner(download *HTTPDownload) {
	for t.progressSpinnerRunning {
		t.t.Logf("[%s] %s/%s @ %sps", download.Name(), humanize.Bytes(uint64(download.CompletedLength())), humanize.Bytes(uint64(download.TotalLength())), humanize.Bytes(uint64(download.Speed())))
		time.Sleep(1 * time.Second)
	}
}

func (t *TestHTTPDownloadListener) OnDownloadStart(dl *HTTPDownload) {
	t.startTime = time.Now()
	t.progressSpinnerRunning = true
	go t.progressSpinner(dl)
}

func (t *TestHTTPDownloadListener) OnDownloadStop(dl *HTTPDownload) {
	t.exitChan <- struct{}{}
}

func (t *TestHTTPDownloadListener) OnDownloadComplete(dl *HTTPDownload) {
	t.progressSpinnerRunning = false
	t.t.Logf("Download took %s", time.Since(t.startTime).String())
	t.exitChan <- struct{}{}
}

func TestFallocateDownload(t *testing.T) {
	var exit chan struct{} = make(chan struct{})
	downloader := NewHTTPDownloader(&http.Client{})
	downloader.AddListener(&TestHTTPDownloadListener{t: t, exitChan: exit})
	_, err := downloader.AddDownload("https://speedtest-ny.turnkeyinternet.net/1000mb.bin", &AddDownloadOpts{
		Connections: 10,
		Dir:         t.TempDir(),
		Fallocate:   true,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	<-exit
}

func TestNormalDownload(t *testing.T) {
	var exit chan struct{} = make(chan struct{})
	downloader := NewHTTPDownloader(&http.Client{})
	downloader.AddListener(&TestHTTPDownloadListener{t: t, exitChan: exit})
	_, err := downloader.AddDownload("https://speedtest-ny.turnkeyinternet.net/1000mb.bin", &AddDownloadOpts{
		Connections: 10,
		Dir:         t.TempDir(),
		Fallocate:   false,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	<-exit
}
