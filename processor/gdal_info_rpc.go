package processor

import (
	"../rpcflow"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type GDALInfoRPC struct {
	In    chan string
	Out   chan rpcflow.GDALFile
	Error chan error
	Port  int
}

func NewGDALInfoRPC(port int, errChan chan error) *GDALInfoRPC {
	return &GDALInfoRPC{
		In:    make(chan string, 100),
		Out:   make(chan rpcflow.GDALFile, 100),
		Error: errChan,
		Port:  port,
	}
}

func (gi *GDALInfoRPC) Run() {
	defer close(gi.Out)

	cmd := exec.Command("../rpc_server/gdalinfo", "-p", fmt.Sprintf("%d", gi.Port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	time.Sleep(500 * time.Millisecond)
	client := rpcflow.NewGDALInfoClient(gi.Port)

	for path := range gi.In {
		res, err := client.Extract(path)
		if err != nil {
			gi.Error <- err
			continue
		}
		gi.Out <- res
	}
}
