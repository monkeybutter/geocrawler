package main

import (
	proc "../processor"
	"../rpcflow"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
	re := flag.String("re", ".*", "Regular expression to match specific files when crawling.")
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	contains := regexp.MustCompile(*re)
	fmt.Println(contains)

	crawlPath, _ := filepath.Abs(flag.Arg(0))

	errChan := make(chan error)
	go func() {
		for err := range errChan {
			log.Println(err)
		}
	}()

	//fc := proc.NewFileCrawler(crawlPath, contains, errChan)
	fc := proc.NewS3Crawler(crawlPath, errChan)
	//pi := proc.NewPosixInfo(errChan)
	gi1 := proc.NewGDALInfoRPC(1234, errChan)
	gi2 := proc.NewGDALInfoRPC(1235, errChan)
	gi3 := proc.NewGDALInfoRPC(1236, errChan)
	gi4 := proc.NewGDALInfoRPC(1237, errChan)
	gp := proc.NewGeoParser(errChan)
	jp := proc.NewJSONPrinter(errChan)

	gi1.In = fc.Out
	gi2.In = fc.Out
	gi3.In = fc.Out
	gi4.In = fc.Out
	giOut := mergeGDALInfoRPCChans(gi1.Out, gi2.Out, gi3.Out, gi4.Out)
	gp.In = giOut
	jp.In = gp.Out

	go fc.Run()
	go gi1.Run()
	go gi2.Run()
	go gi3.Run()
	go gi4.Run()
	go gp.Run()
	jp.Run()

}
