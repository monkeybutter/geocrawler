package main

import (
	proc "../processor"
	"../rpcflow"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func mergeGDALInfoRPCChans(cs ...chan rpcflow.GDALFile) chan rpcflow.GDALFile {
	var wg sync.WaitGroup
	out := make(chan rpcflow.GDALFile)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c chan rpcflow.GDALFile) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	crawlPath, _ := filepath.Abs(flag.Arg(0))

	errChan := make(chan error)
	go func() {
		for err := range errChan {
			log.Println(err)
		}
	}()

	fc := proc.NewFileCrawler(crawlPath, errChan)
	ss := proc.NewStringSplitter(errChan)
	pi := proc.NewPosixInfo(errChan)
	gi1 := proc.NewGDALInfoRPC(1234, errChan)
	gi2 := proc.NewGDALInfoRPC(1235, errChan)
	gi3 := proc.NewGDALInfoRPC(1236, errChan)
	gi4 := proc.NewGDALInfoRPC(1237, errChan)
	gp := proc.NewGeoParser(errChan)
	//gi := proc.NewGDALInfo(errChan)

	ss.In = fc.Out
	gi1.In = ss.Out1
	gi2.In = ss.Out1
	gi3.In = ss.Out1
	gi4.In = ss.Out1
	giOut := mergeGDALInfoRPCChans(gi1.Out, gi2.Out, gi3.Out, gi4.Out)
	gp.In = giOut
	pi.In = ss.Out2
	//gi2.Out = gi1.Out
	//gi3.Out = gi1.Out
	//gi4.Out = gi1.Out

	go fc.Run()
	go ss.Run()
	go gi1.Run()
	//go gi2.Run()
	//go gi3.Run()
	//go gi4.Run()
	go gp.Run()
	pi.Run()
	//go gi2.Run()
	//go gi3.Run()
	//go gi4.Run()

}
