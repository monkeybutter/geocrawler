package main

import (
	"../geolib"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func main() {

	concLevel := flag.Int("c", 1, "Concurrency level.")
	re := flag.String("re", ".*", "Regular expression to match specific files when crawling.")

	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	contains := regexp.MustCompile(*re)

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

		filePaths := geolib.FilesProducerSerial(crawlPath)

		workers := geolib.NewConcLimiter(*concLevel)
		for path := range filePaths {
			if contains.MatchString(path) {
				workers.Increase()
				go geolib.GDALMetadataPrinter(path, workers)
			}
		}

		workers.Wait()

	} else {
		out, err := geolib.GetGDALMetadata(crawlPath)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	}

}
