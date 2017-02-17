package rpcflow

import (
	"fmt"
	"log"
	"net/rpc"
)

type Args struct {
	FilePath string
}

type Overlay struct {
	Index int
	XSize int
	YSize int
}

type GDALDataSet struct {
	DataSetName  string
	RasterCount  int
	Type         string
	XSize        int
	YSize        int
	ProjWKT      string
	GeoTransform []float64
	Overlays     []Overlay
	Extras       map[string][]string
}

type GDALFile struct {
	FileName string
	Driver   string
	DataSets []GDALDataSet
}

type GDALInfoClient struct {
	client *rpc.Client
}

func NewGDALInfoClient(port int) *GDALInfoClient {
	c, err := rpc.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatal("dialing:", err)
	}

	return &GDALInfoClient{client: c}
}

func (t *GDALInfoClient) Extract(filePath string) (GDALFile, error) {
	args := &Args{filePath}
	var reply GDALFile
	err := t.client.Call("GDALInfo.Extract", args, &reply)

	return reply, err
}
