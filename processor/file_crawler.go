package processor

import (
	"os"
	"path/filepath"
)

type FileCrawler struct {
	Out   chan string
	Error chan error
	root  string
}

func NewFileCrawler(rootPath string, errChan chan error) *FileCrawler {
	return &FileCrawler{
		Out:   make(chan string, 100),
		Error: errChan,
		root:  rootPath,
	}
}

func (fc *FileCrawler) Run() {
	defer close(fc.Out)

	fInfo, err := os.Stat(fc.root)
	if err != nil {
		fc.Error <- err
		return
	}

	if fInfo.IsDir() {
		filepath.Walk(fc.root, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				fc.Out <- path
			}
			return nil
		})
	} else {
		fc.Out <- fc.root
	}
}
