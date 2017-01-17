package main

import (
	proc "../processor"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	//"regexp"
)

func main() {

	//concLevel := flag.Int("c", 1, "Concurrency level.")
	//re := flag.String("re", ".*", "Regular expression to match specific files when crawling.")

	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	//contains := regexp.MustCompile(*re)

	crawlPath, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	errChan := make(chan error)

	go func() {
		for range errChan {
			continue
		}
	}()

	fc := proc.NewFileCrawler(crawlPath, errChan)
	gi1 := proc.NewGDALInfo(errChan)
	gi2 := proc.NewGDALInfo(errChan)
	gt := proc.NewGeoTransformer(errChan)
	jp := proc.NewJSONPrinter(errChan)

	gi1.In = fc.Out
	gi2.In = fc.Out
	gi2.Out = gi1.Out
	gt.In = gi1.Out
	jp.In = gt.Out

	go fc.Run()
	go gi1.Run()
	go gi2.Run()
	go gt.Run()
	jp.Run()

	for value := range(gi1.Out) {
		fmt.Println(value)
	}

}
