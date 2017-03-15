package processor

import (
	"../rpcflow"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type GDALInfoRPCS3 struct {
	In    chan string
	Out   chan rpcflow.GDALFile
	Error chan error
	Port  int
}

func NewGDALInfoRPCS3(port int, errChan chan error) *GDALInfoRPC {
	return &GDALInfoRPC{
		In:    make(chan string, 100),
		Out:   make(chan rpcflow.GDALFile, 100),
		Error: errChan,
		Port:  port,
	}
}

func (gi *GDALInfoRPCS3) Run() {
	defer close(gi.Out)
	var client *rpcflow.GDALInfoClient
	restart := make(chan struct{})

	go func() {
		for true {
			cmd := exec.Command("../rpc_server/gdalinfo", "-p", fmt.Sprintf("%d", gi.Port))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Start()
			if err != nil {
				panic(err)
			}
			time.Sleep(500 * time.Millisecond)
			client = rpcflow.NewGDALInfoClient(gi.Port)
			<-restart
			if err := cmd.Process.Kill(); err != nil {
				panic(err)
			}
			cmd.Wait()
		}
	}()

	time.Sleep(time.Second)
	i := 0
	for path := range gi.In {
		if i > 1000 {
			i = 0
			restart <- struct{}{}
			time.Sleep(time.Second)
		}
		res, err := client.Extract(path)
		if err != nil {
			gi.Error <- err
			continue
		}
		gi.Out <- res
		i++
	}
}
