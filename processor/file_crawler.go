package processor

import (
	pb "../grpc/gdalservice"
	"os"
	"path/filepath"
	"regexp"
)

type FileCrawler struct {
	In    chan string
	Out   chan *pb.GeoRequest
	Error chan error
	root  string
	re    *regexp.Regexp
}

func NewFileCrawler(rootPath string, contains *regexp.Regexp, errChan chan error) *FileCrawler {
	return &FileCrawler{
		In:    make(chan string, 100),
		Out:   make(chan *pb.GeoRequest, 100),
		Error: errChan,
		root:  rootPath,
		re:    contains,
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
			if !info.IsDir() && fc.re.MatchString(path) && filepath.Ext(path) != ".ovr" {
				fc.Out <- &pb.GeoRequest{path}
			}
			return nil
		})
	} else {
		fc.Out <- &pb.GeoRequest{fc.root}
	}
}
