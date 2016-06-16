package main

import (
	"fmt"
	"flag"
	"path/filepath"
	"os"
	"./geocrawl"
)

//env GOOS=darwin GOARCH=amd64 go build test.go

func main() {

	concLevel := flag.Int("c", 1, "Concurrency level.")

	flag.Parse()

	if len(flag.Args()) != 1 {
    		flag.Usage()
    		os.Exit(1)
	}

	crawlPath, err := filepath.Abs(flag.Arg(0))

	if err != nil {
		fmt.Println("Error:", err)
    		os.Exit(1)
	}

	fInfo, err := os.Stat(crawlPath)

	if err != nil {
		fmt.Println("Error:", err)
    	os.Exit(1)
	}

	if fInfo.IsDir() {

		filePaths := geocrawl.FilesProducerSerial(crawlPath)

		workers := geocrawl.NewConcLimiter(*concLevel)
		for path := range filePaths {
			workers.Increase()
			go geocrawl.GDALMetadataPrinter(path, workers)
		}

		workers.Wait()

	} else {
		out, err := geocrawl.GetGDALMetadata(crawlPath)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	}	

	
	
	


	


}
