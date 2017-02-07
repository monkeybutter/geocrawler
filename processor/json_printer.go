package processor

import (
	"encoding/json"
	"fmt"
)

type JSONPrinter struct {
	In    chan GeoFile
	Error chan error
}

func NewJSONPrinter(errChan chan error) *JSONPrinter {
	return &JSONPrinter{
		In:    make(chan GeoFile),
		Error: errChan,
	}
}

func (jp *JSONPrinter) Run() {

	for geoFile := range jp.In {
		out, err := json.Marshal(&geoFile)
		if err != nil {
			jp.Error <- err
			return
		}
		fmt.Printf("%s\tgdal\t%s\n", geoFile.FileName, string(out))
	}
}
