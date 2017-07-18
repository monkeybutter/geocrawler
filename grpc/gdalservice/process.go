package gdalservice

import (
	"context"
	//"log"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"bytes"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
)

type ErrorMsg struct {
	Address string
	Replace bool
	Error   error
}

func createUniqueFilename(dir string) (string, error) {
	var filename string
	var err error

	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			return filename, fmt.Errorf("Could not get the current working directory %v", err)
		}
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return filename, fmt.Errorf("The provided directory does not exist %v", err)
	}

	for {
		filename = filepath.Join(dir, strconv.FormatUint(rand.Uint64(), 16))
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return filename, nil
		}
	}
}

type InfoTask struct {
	Payload *GeoRequest
	Resp    chan *GeoFile
	Error   chan error
}

type Process struct {
	Context       context.Context
	CancelFunc    context.CancelFunc
	InfoTaskQueue chan *InfoTask
	InfoCmd       *exec.Cmd
	InfoAddress   string
	ErrorMsg      chan *ErrorMsg
}

func NewProcess(ctx context.Context, infoTQueue chan *InfoTask, infoBinary string, errChan chan *ErrorMsg) *Process {
	newCtx, cancel := context.WithCancel(ctx)
	infoAddr, err := createUniqueFilename("/tmp")
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(infoAddr); !os.IsNotExist(err) {
		os.Remove(infoAddr)
	}

	return &Process{newCtx, cancel, infoTQueue, exec.CommandContext(newCtx, infoBinary, infoAddr), infoAddr, errChan}
}

func (p *Process) waitInfoReady() error {
	timer := time.NewTimer(time.Millisecond * 200)
	ready := make(chan struct{})

	go func(signal chan struct{}) {
		for {
			time.Sleep(time.Millisecond * 20)
			if _, err := os.Stat(p.InfoAddress); err == nil {
				signal <- struct{}{}
				return

			}
		}
	}(ready)

	select {
	case <-ready:
		return nil
	case <-timer.C:
		return fmt.Errorf("Address file creation timed out")
	}
}

func (p *Process) Start() {
	err := p.InfoCmd.Start()
	if err != nil {
		p.ErrorMsg <- &ErrorMsg{p.InfoAddress, true, err}
		return
	}

	err = p.waitInfoReady()
	if err != nil {
		p.ErrorMsg <- &ErrorMsg{p.InfoAddress, true, err}
		return
	}

	log.Printf("Process running with PID %d\n", p.InfoCmd.Process.Pid)

	go func() {
		for iTask := range p.InfoTaskQueue {

			conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: p.InfoAddress, Net: "unix"})
			if err != nil {
				iTask.Error <- fmt.Errorf("dial failed: %v", err)
				p.ErrorMsg <- &ErrorMsg{p.InfoAddress, true, err}
				break
			}

			inb, err := proto.Marshal(iTask.Payload)
			if err != nil {
				conn.Close()
				iTask.Error <- fmt.Errorf("encode failed: %v", err)
				continue
			}

			n, err := conn.Write(inb)
			if err != nil {
				conn.Close()
				iTask.Error <- fmt.Errorf("error writing %d bytes of data: %v", n, err)
				continue
			}
			conn.CloseWrite()

			var buf bytes.Buffer
			nr, err := io.Copy(&buf, conn)
			if err != nil {
				conn.Close()
				iTask.Error <- fmt.Errorf("error reading %d bytes of data: %v", nr, err)
				continue
			}
			conn.Close()

			out := new(GeoFileAndError)
			err = proto.Unmarshal(buf.Bytes(), out)
			if err != nil {
				iTask.Error <- fmt.Errorf("error decoding data: %v %s %d", err, iTask.Payload, n)
				continue
			}

			if out.Error != "OK" {
				iTask.Error <- fmt.Errorf("Warp Process %s reported: %s", p.InfoAddress, out.Error)
				continue

			}

			iTask.Resp <- out.GeoFile

		}
	}()

	go func() {
		p.InfoCmd.Wait()
		err = os.Remove(p.InfoAddress)
		if err != nil {
			p.ErrorMsg <- &ErrorMsg{p.InfoAddress, false, fmt.Errorf("Couldn't delete unix connection file")}
			return
		}

		select {
		case <-p.Context.Done():
			p.ErrorMsg <- &ErrorMsg{p.InfoAddress, false, p.Context.Err()}
		default:
			p.ErrorMsg <- &ErrorMsg{p.InfoAddress, true, fmt.Errorf("Process finished unexpectedly")}
		}
	}()
}

/*
func (p *Process) Cancel() {
	p.CancelFunc()
	time.Sleep(100 * time.Millisecond)
}
*/
