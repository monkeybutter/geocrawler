package gdalservice

import (
	"context"
	"fmt"
	"log"
)

type ProcessPool struct {
	Pool      []*Process
	InfoTaskQueue chan *InfoTask
	ErrorMsg     chan *ErrorMsg
}

func (p *ProcessPool) AddInfoQueue(task *InfoTask) {
	if len(p.InfoTaskQueue) > 390 {
		task.Error <- fmt.Errorf("Pool TaskQueue is full")
		return
	}
	p.InfoTaskQueue <- task
}

//func (p *ProcessPool) AddProcess(errChan chan error, healthChan chan *HealthMsg) {
func (p *ProcessPool) AddProcess() {
	proc := NewProcess(context.Background(), p.InfoTaskQueue, "./gdalinfo_process", p.ErrorMsg)
	proc.Start()
	p.Pool = append(p.Pool, proc)
}

/*
func (p *ProcessPool) RemoveProcess(address string) {
	newPool := []*Process{}
	for _, proc := range p.Pool {
		if proc.Address != address {
			newPool = append(newPool, proc)
		}
	}
	p.Pool = newPool
}
*/

func CreateProcessPool(n int) *ProcessPool {
	p := &ProcessPool{[]*Process{}, make(chan *InfoTask, 400), make(chan *ErrorMsg)}

	go func() {
		for {
			select {
			case err := <-p.ErrorMsg:
				log.Println("Process needs to be restarded?", err)
			}
		}
	}()

	for i := 0; i < n; i++ {
		p.AddProcess()
	}


	/*
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		for {
			select {
			case <-signals:
				p.DeleteProcessPool()
				time.Sleep(1 * time.Second)
				os.Exit(1)
			}
		}
	}()
	*/


	/*
	go func() {
		for {
			select {
			case hMsg := <-p.Health:
				p.RemoveProcess(hMsg.Address)
				if hMsg.Replace == true {
					p.AddProcess(p.Error, p.Health)
				}
			}
		}
	}()
	*/

	return p
}

/*
func (p *ProcessPool) DeleteProcessPool() {
	for _, proc := range p.Pool {
		proc.Cancel()
	}
}
*/
