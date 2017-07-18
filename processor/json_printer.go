package processor

import (
	"encoding/json"
	"fmt"
	pb "../grpc/gdalservice"
	"github.com/golang/protobuf/ptypes/struct"
)

type JSONPrinter struct {
	In    chan *pb.GeoFile
	Out   chan struct{}
	Error chan error
}

func NewJSONPrinter(errChan chan error) *JSONPrinter {
	return &JSONPrinter{
		In:    make(chan *pb.GeoFile, 100),
		Out:   make(chan struct{}),
		Error: errChan,
	}
}

func (jp *JSONPrinter) Run() {
	defer close(jp.Out)

	for geoFile := range jp.In {
		out, err := json.Marshal(&geoFile)
		if err != nil {
			jp.Error <- err
			fmt.Println(string(out))
			return
		}
		fmt.Printf("%s\tgdal\t%s\n", geoFile.FileName, string(out))
	}

}
