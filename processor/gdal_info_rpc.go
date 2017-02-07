package processor

import (
	"../rpcflow"
)

type GDALInfoRPC struct {
	In     chan string
	Out    chan rpcflow.GDALFile
	Error  chan error
	Client *rpcflow.GDALInfoClient
}

func NewGDALInfoRPC(port int, errChan chan error) *GDALInfoRPC {
	return &GDALInfoRPC{
		In:     make(chan string, 100),
		Out:    make(chan rpcflow.GDALFile, 100),
		Error:  errChan,
		Client: rpcflow.NewGDALInfoClient(port),
	}
}

func (gi *GDALInfoRPC) Run() {
	defer close(gi.Out)

	for path := range gi.In {
		res, err := gi.Client.Extract(path)
		if err != nil {
			gi.Error <- err
			continue
		}
		gi.Out <- res
	}
}
