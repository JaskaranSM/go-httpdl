package httpdl

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
}

func (d *DownloadPart) CompletedLength() int64 {
	return d.completed
}

func (d *DownloadPart) TotalLength() int64 {
	return d.total
}

func (d *DownloadPart) HandleResponse(resp *http.Response) error {
	buffer := make([]byte, d.chunksize)

	for {
		if d.isCancelled || d.isFailed || d.isCompleted {
			return nil
		}

		nbytes, err := resp.Body.Read(buffer[0:d.chunksize])
		if err != nil && err != io.EOF {
			d.isFailed = true
			d.Err = err
			return err
		}
		nbytes, err = d.file.WriteAt(buffer[0:nbytes], int64(d.offset)+d.completed)
		if err != nil {
			d.isFailed = true
			d.Err = err
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
	log.Println(rangeHeader)
	req.Header.Set("Range", rangeHeader)
	resp, err := d.client.Do(req)
	if err != nil {
		d.Err = err
		d.isFailed = true
		return err
	}
	defer resp.Body.Close()
	return d.HandleResponse(resp)
}
