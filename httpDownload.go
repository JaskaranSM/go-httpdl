package httpdl

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

type HTTPDownload struct {
	gid         string
	url         string
	connections int
	size        int64
	dir         string
	file        *os.File
	client      *http.Client
	parts       []*DownloadPart
	isCancelled bool
	speed       int64
	chunksize   int64
	name        string
}

func (h *HTTPDownload) Init() {
	if h.size == 4096*1024 {
		h.connections = 1
	}
	sizePerPart := h.size / int64(h.connections)
	for i := 0; i < h.connections; i++ {
		part := &DownloadPart{
			Id:     i,
			url:    h.url,
			offset: i * int(sizePerPart),
		}
		switch {
		case i == h.connections-1:
			part.total = h.size - sizePerPart*int64(i)
			break
		default:
			part.total = sizePerPart
		}
		part.completed = 0
		part.file = h.file
		part.client = h.client
		part.chunksize = h.chunksize
		h.parts = append(h.parts, part)
	}
}

func (h *HTTPDownload) IsCompleted() bool {
	if h.IsCancelled() {
		return true
	}
	if h.IsFailed() {
		return true
	}

	for _, part := range h.parts {
		if !part.isCompleted {
			return false
		}
	}
	return true
}

func (h *HTTPDownload) IsCancelled() bool {
	for _, part := range h.parts {
		if !part.isCancelled {
			return false
		}
	}
	return true
}

func (h *HTTPDownload) IsFailed() bool {
	for _, part := range h.parts {
		if !part.isFailed {
			return false
		}
	}
	return true
}

func (h *HTTPDownload) GetFailureError() error {
	part := h.getFailedPart()
	if part == nil {
		return nil
	}
	return part.Err
}

func (h *HTTPDownload) getFailedPart() *DownloadPart {
	for _, part := range h.parts {
		if part.isFailed {
			return part
		}
	}
	return nil
}

func (h *HTTPDownload) GetFileHandle() *os.File {
	return h.file
}

func (h *HTTPDownload) SpeedObserver() {
	var last int64
	for range time.Tick(1 * time.Second) {
		if h.isCancelled || h.IsCompleted() {
			return
		}
		new := h.CompletedLength()
		h.speed = new - last
		last = new
	}
}

func (h *HTTPDownload) Name() string {
	return h.name
}

func (h *HTTPDownload) CompletedLength() int64 {
	var completed int64
	for i := 0; i < h.connections; i++ {
		completed += h.parts[i].CompletedLength()
	}
	return completed
}

func (h *HTTPDownload) Speed() int64 {
	return h.speed
}

func (h *HTTPDownload) TotalLength() int64 {
	return h.size
}

func (h *HTTPDownload) Gid() string {
	return h.gid
}

func (h *HTTPDownload) StartDownload() {
	for i := 0; i < h.connections; i++ {
		go h.parts[i].Download()
	}
	go h.SpeedObserver()
}

func (h *HTTPDownload) CancelDownload() {
	h.isCancelled = true
	for i := 0; i < h.connections; i++ {
		h.parts[i].isCancelled = true
		h.parts[i].Err = fmt.Errorf("Cancelled by user.")
	}
}
