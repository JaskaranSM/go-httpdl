package httpdl

type DownloadEvent int

const (
	OnDownloadStartEvent    = 0
	OnDownloadCompleteEvent = 1
	OnDownloadStopEvent     = 2
)

type HTTPDownloadListener interface {
	OnDownloadStart(*HTTPDownload)
	OnDownloadComplete(*HTTPDownload)
	OnDownloadStop(*HTTPDownload)
}
