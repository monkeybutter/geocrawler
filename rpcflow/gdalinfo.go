package rpcflow

import (
	"fmt"
	"log"
	"net/rpc"
)

type Args struct {
	FilePath string
}

type GDALDataSet struct {
	DataSetName  string
	RasterCount  int
	Type         string
	XSize        int
	YSize        int
	ProjWKT      string
	GeoTransform []float64
	Extras       map[string][]string
}

type Result struct {
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

func (t *GDALInfoClient) Extract(filePath string) (Result, error) {
	args := &Args{filePath}
	var reply Result
	err := t.client.Call("GDALInfo.Extract", args, &reply)
	
	return reply, err
}
