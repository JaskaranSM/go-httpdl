package httpdl

type AddDownloadOpts struct {
	Connections int
	Dir         string
	Chunksize   int64
}

type URLProperties struct {
	SupportsMultiConnection bool
	Size                    int64
	SupportsRange           bool
	Filename                string
}
