package httpdl

import (
	"mime"
	"net/http"
	u "net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func NewHTTPDownloader(client *http.Client) *HTTPDownloader {
	return &HTTPDownloader{
		client: client,
	}
}

type HTTPDownloader struct {
	listeners []HTTPDownloadListener
	downloads []*HTTPDownload
	client    *http.Client
}

func (h *HTTPDownloader) GetURLProperties(url string) (*URLProperties, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sizeInt64, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	props := &URLProperties{}
	props.SupportsMultiConnection = h.HeaderSupportsByteRange(resp.Header)
	props.SupportsRange = h.HeaderSupportsByteRange(resp.Header)
	props.Filename = h.SniffFilename(url, resp.Header)
	props.Size = sizeInt64
	return props, nil
}

func (h *HTTPDownloader) NotifyListeners(event DownloadEvent, download *HTTPDownload) {
	for _, listener := range h.listeners {
		switch event {
		case OnDownloadStartEvent:
			listener.OnDownloadStart(download)
		case OnDownloadStopEvent:
			listener.OnDownloadStop(download)
		case OnDownloadCompleteEvent:
			listener.OnDownloadComplete(download)
		}
	}
}

func (h *HTTPDownloader) ListenForEvents(download *HTTPDownload) {
	for {
		if download.IsCompleted() {
			h.NotifyListeners(OnDownloadCompleteEvent, download)
			break
		}
		if download.IsCancelled() {
			h.NotifyListeners(OnDownloadStopEvent, download)
			break
		}
		if download.IsFailed() {
			h.NotifyListeners(OnDownloadStopEvent, download)
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func (h *HTTPDownloader) AddListener(listener HTTPDownloadListener) {
	h.listeners = append(h.listeners, listener)
}

func (h *HTTPDownloader) AddDownload(url string, opts *AddDownloadOpts) (*HTTPDownload, error) {
	props, err := h.GetURLProperties(url)
	if err != nil {
		return nil, err
	}
	if !props.SupportsMultiConnection || props.Size == 0 {
		opts.Connections = 1
	}
	if opts.Chunksize <= 0 {
		opts.Chunksize = 4096
	}
	if opts.Filename != "" {
		props.Filename = opts.Filename
	}
	if opts.Size > 0 {
		props.Size = opts.Size
	}
	pth := path.Join(opts.Dir, props.Filename)
	if _, err := os.Stat(pth); err == nil {
		os.Remove(pth)
	}
	os.MkdirAll(opts.Dir, 0755)
	file, err := os.OpenFile(pth, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	download := &HTTPDownload{
		gid:         RandString(16),
		url:         url,
		connections: opts.Connections,
		size:        props.Size,
		dir:         opts.Dir,
		file:        file,
		client:      h.client,
		isCancelled: false,
		speed:       0,
		chunksize:   opts.Chunksize,
		name:        props.Filename,
	}
	download.Init()
	h.downloads = append(h.downloads, download)
	download.StartDownload()
	h.NotifyListeners(OnDownloadStartEvent, download)
	go h.ListenForEvents(download)
	return download, nil
}

func (h *HTTPDownloader) HeaderSupportsByteRange(header http.Header) bool {
	acceptRange := header.Get("Accept-Ranges")
	if acceptRange == "" {
		return false
	}
	return strings.Contains(acceptRange, "bytes")
}

func (h *HTTPDownloader) SniffFilename(url string, header http.Header) string {
	var filename string
	var err error
	_, params, _ := mime.ParseMediaType(header.Get("Content-Disposition"))
	filename = params["filename"]
	if filename == "" {
		d := strings.Split(url, "/")
		filename, err = u.QueryUnescape(d[len(d)-1])
		if err != nil {
			filename = d[len(d)-1]
		}
	}
	return filename
}
