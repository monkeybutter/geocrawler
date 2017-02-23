package main

import (
	proc "../processor"
	"../rpcflow"
	"flag"
	"log"
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
	a := flag.String("a", "None", "After after a certain key.")
	flag.Parse()

	if *a == "None" {
		a = nil
	}
	
	errChan := make(chan error)
	go func() {
		for err := range errChan {
			log.Println(err)
		}
	}()

	fc := proc.NewS3Crawler(a, errChan)
	gi1 := proc.NewGDALInfoRPC(1234, errChan)
	gi2 := proc.NewGDALInfoRPC(1235, errChan)
	gi3 := proc.NewGDALInfoRPC(1236, errChan)
	gi4 := proc.NewGDALInfoRPC(1237, errChan)
	gi5 := proc.NewGDALInfoRPC(1238, errChan)
	gi6 := proc.NewGDALInfoRPC(1239, errChan)
	gi7 := proc.NewGDALInfoRPC(1240, errChan)
	gi8 := proc.NewGDALInfoRPC(1241, errChan)
	gp := proc.NewGeoParser(errChan)
	jp := proc.NewJSONPrinter(errChan)

	gi1.In = fc.Out
	gi2.In = fc.Out
	gi3.In = fc.Out
	gi4.In = fc.Out
	gi5.In = fc.Out
	gi6.In = fc.Out
	gi7.In = fc.Out
	gi8.In = fc.Out
	giOut := mergeGDALInfoRPCChans(gi1.Out, gi2.Out, gi3.Out, gi4.Out, gi5.Out, gi6.Out, gi7.Out, gi8.Out)
	gp.In = giOut
	jp.In = gp.Out

	go fc.Run()
	go gi1.Run()
	go gi2.Run()
	go gi3.Run()
	go gi4.Run()
	go gi5.Run()
	go gi6.Run()
	go gi7.Run()
	go gi8.Run()
	go gp.Run()
	jp.Run()
}
