package main

import (
	proc "./processor"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

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
			fmt.Println(err)
		}
	}()

	fc := proc.NewFileCrawler(crawlPath, errChan)
	gi1 := proc.NewGDALInfoRPC(1234, errChan)
	gi2 := proc.NewGDALInfoRPC(1235, errChan)
	gi3 := proc.NewGDALInfoRPC(1236, errChan)
	gi4 := proc.NewGDALInfoRPC(1237, errChan)
	//gi := proc.NewGDALInfo(errChan)

	gi1.In = fc.Out
	gi2.In = fc.Out
	gi3.In = fc.Out
	gi4.In = fc.Out
	gi2.Out = gi1.Out
	gi3.Out = gi1.Out
	gi4.Out = gi1.Out

	go fc.Run()
	go gi1.Run()
	go gi2.Run()
	go gi3.Run()
	go gi4.Run()

	for val := range gi1.Out {
		fmt.Println(val)
	}
}
