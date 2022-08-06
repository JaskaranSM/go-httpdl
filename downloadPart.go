package httpdl

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

type DownloadPart struct {
	Id          int
	url         string
	offset      int
	isCancelled bool
	isCompleted bool
	isFailed    bool
	Err         error
	completed   int64
	total       int64
	chunksize   int64
	file        *os.File
	client      *http.Client
	mut         sync.Mutex
}

func (d *DownloadPart) CompletedLength() int64 {
	return d.completed
}

func (d *DownloadPart) TotalLength() int64 {
	return d.total
}

func (d *DownloadPart) Write(p []byte) (int, error) {
	if d.isCancelled {
		return 0, fmt.Errorf("Cancelled by user.")
	}
	d.completed += int64(len(p))
	return d.file.Write(p)
}

func (d *DownloadPart) HandleResponseWriter(resp *http.Response) error {
	_, err := io.Copy(d, resp.Body)
	if err != nil && err != io.EOF {
		d.mut.Lock()
		defer d.mut.Unlock()
		d.Err = err
		d.isFailed = true
		return err
	}
	if d.total == 0 {
		d.total = d.completed
	}
	d.isCompleted = true
	return nil
}

func (d *DownloadPart) HandleResponse(resp *http.Response) error {
	buffer := make([]byte, d.chunksize)

	for {
		if d.isCancelled || d.isFailed || d.isCompleted {
			return nil
		}
		nbytes, err := resp.Body.Read(buffer[0:d.chunksize])
		if err != nil && err != io.EOF {
			d.mut.Lock()
			defer d.mut.Unlock()
			d.Err = err
			d.isFailed = true
			return err
		}
		nbytes, err = d.file.WriteAt(buffer[0:nbytes], int64(d.offset)+d.completed)
		if err != nil {
			d.mut.Lock()
			defer d.mut.Unlock()
			d.Err = err
			d.isFailed = true
			return nil
		}
		d.completed += int64(nbytes)
		remaining := d.total - d.completed
		switch {
		case remaining == 0:
			d.isCompleted = true
			return nil
		case remaining < d.chunksize:
			d.chunksize = d.total - d.completed
		}
	}
}

func (d *DownloadPart) Download() error {
	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		return err
	}
	rangeHeader := fmt.Sprintf("bytes=%d-%d", d.offset, int64(d.offset)+d.total-1)
	req.Header.Set("Range", rangeHeader)
	resp, err := d.client.Do(req)
	if err != nil {
		d.mut.Lock()
		defer d.mut.Unlock()
		d.Err = err
		d.isFailed = true
		return err
	}
	defer resp.Body.Close()
	if d.total == 0 {
		return d.HandleResponseWriter(resp)
	}
	return d.HandleResponse(resp)
}
