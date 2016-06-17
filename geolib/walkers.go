package geolib

import (
	"path/filepath"
	"os"
	"io/ioutil"
	"fmt"
)

func walkDir(dir string, lim *ConcLimiter, filePaths chan<- string) {
        defer lim.Decrease()

        for _, entry := range listDir(dir) {
                if entry.IsDir() {
                        lim.Increase()
                        go walkDir(filepath.Join(dir, entry.Name()), lim, filePaths)
                } else {
                        filePaths <- filepath.Join(dir, entry.Name())
                }
        }
}

// returns the entries of directory dir.
func listDir(dir string) []os.FileInfo {
        entries, err := ioutil.ReadDir(dir)
        if err != nil {
                fmt.Fprintf(os.Stderr, "du: %v\n", err)
                return nil
        }
        return entries
}

func FilesProducer(root string, concLevel int) chan string {
        filePaths := make(chan string)

        lim := NewConcLimiter(concLevel)

        lim.Increase()
        go walkDir(root, lim, filePaths)

        go func() {
                lim.Wait()
                close(filePaths)
        }()

        return filePaths
}

func FilesProducerSerial(root string) chan string {
        filePaths := make(chan string)

        go func() {
                filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
                        if !info.IsDir() {
                                filePaths <- path
                        }
                        return nil
                })
                close(filePaths)
        }()

        return filePaths
}
