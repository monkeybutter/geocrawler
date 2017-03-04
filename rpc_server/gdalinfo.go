package main

// #include <stdio.h>
// #include <stdlib.h>
// #include "gdal.h"
// #include "ogr_srs_api.h" /* for SRS calls */
// #include "cpl_string.h"
// #cgo LDFLAGS: -lgdal
//void openFile(char *cPath)
//{
//	GDALDatasetH hDataset;
//	hDataset = GDALOpen(cPath, GDAL_OF_READONLY);
//	GDALClose(hDataset);
//	return; 
//}
import "C"

import (
	"../rpcflow"
	"flag"
	"fmt"
	"math"
	"net"
	"net/rpc"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var dateFormats []string = []string{"2006-01-02 15:04:05.0", "2006-1-2 15:4:5"}

var durationUnits map[string]time.Duration = map[string]time.Duration{"seconds": time.Second, "hours": time.Hour, "days": time.Hour * 24}

var CWGS84WKT *C.char = C.CString(`GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],TOWGS84[0,0,0,0,0,0,0],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.017453292        5199433,AUTHORITY["EPSG","9108"]],AUTHORITY["EPSG","4326"]]","proj4":"+proj=longlat +ellps=WGS84 +towgs84=0,0,0,0,0,0,0 +no_defs `)

var CsubDS *C.char = C.CString("SUBDATASETS")
var CtimeUnits *C.char = C.CString("time#units")
var CncDimTimeValues *C.char = C.CString("NETCDF_DIM_time_VALUES")

var GDALTypes map[C.GDALDataType]string = map[C.GDALDataType]string{0: "Unkown", 1: "Byte", 2: "Uint16", 3: "Int16",
	4: "UInt32", 5: "Int32", 6: "Float32", 7: "Float64",
	8: "CInt16", 9: "CInt32", 10: "CFloat32", 11: "CFloat64",
	12: "TypeCount"}

type GDALInfo struct{}

func (b *GDALInfo) Extract(args *rpcflow.Args, res *rpcflow.GDALFile) error {
	cPath := C.CString(args.FilePath)
	defer C.free(unsafe.Pointer(cPath))
	hDataset := C.GDALOpenShared(cPath, C.GA_ReadOnly)
	defer C.GDALClose(hDataset)
	
	hDriver := C.GDALGetDatasetDriver(hDataset)
	cShortName := C.GDALGetDriverShortName(hDriver)
	shortName := C.GoString(cShortName)

	metadata := C.GDALGetMetadata(hDataset, CsubDS)
	nsubds := C.CSLCount(metadata) / C.int(2)

	var datasets = []rpcflow.GDALDataSet{}
	if nsubds == C.int(0) {
		// There are no subdatasets
		dsInfo, err := getDataSetInfo(cPath, shortName)
		if err != nil {
			return err
		}
		datasets = append(datasets, dsInfo)

	} else {
		// There are subdatasets
		for i := C.int(1); i <= nsubds; i++ {
			subDSId := C.CString(fmt.Sprintf("SUBDATASET_%d_NAME", i))
			pszSubdatasetName := C.CSLFetchNameValue(metadata, subDSId)
			dsInfo, err := getDataSetInfo(pszSubdatasetName, shortName)
			if err != nil {
				return err
			}
			datasets = append(datasets, dsInfo)
		}
	}

	*res = rpcflow.GDALFile{args.FilePath, shortName, datasets}
	
	return nil
}

func getDataSetInfo(dsName *C.char, driverName string) (rpcflow.GDALDataSet, error) {

	datasetName := C.GoString(dsName)
	hSubdataset := C.GDALOpenShared(dsName, C.GDAL_OF_READONLY)
	if hSubdataset == nil {
		return rpcflow.GDALDataSet{}, fmt.Errorf("GDAL could not open dataset: %s", datasetName)
	}
	defer C.GDALClose(hSubdataset)

	extras := map[string][]string{}
	if driverName == "netCDF" {
		nc_times, err := getNCTime(datasetName, hSubdataset)
		if err == nil && nc_times != nil {
			extras["nc_times"] = nc_times
		}
	}

	hBand := C.GDALGetRasterBand(hSubdataset, 1)
	nOvr := C.GDALGetOverviewCount(hBand)
	ovrs := make([]rpcflow.Overview, int(nOvr))
	for i := C.int(0); i < nOvr; i++ {
		hOvr := C.GDALGetOverview(hBand, i)
		ovrs[int(i)] = rpcflow.Overview{int(C.GDALGetRasterBandXSize(hOvr)), int(C.GDALGetRasterBandYSize(hOvr))}
	}

	pszWkt := C.GDALGetProjectionRef(hSubdataset)
	if C.GoString(pszWkt) == "" {
		pszWkt = CWGS84WKT
	}

	// To hold the geotransform
	dArr := [6]C.double{}
	C.GDALGetGeoTransform(hSubdataset, &dArr[0])
	fArr := (*[6]float64)(unsafe.Pointer(&dArr))
	return rpcflow.GDALDataSet{datasetName, int(C.GDALGetRasterCount(hSubdataset)), GDALTypes[C.GDALGetRasterDataType(hBand)],
		int(C.GDALGetRasterXSize(hSubdataset)), int(C.GDALGetRasterYSize(hSubdataset)), C.GoString(pszWkt), fArr[:], ovrs, extras}, nil

}

func goStrings(argc C.int, argv **C.char) []string {

	length := int(argc)
	tmpslice := (*[1 << 30]*C.char)(unsafe.Pointer(argv))[:length:length]
	gostrings := make([]string, length)
	for i, s := range tmpslice {
		gostrings[i] = C.GoString(s)
	}
	return gostrings
}

func getDate(inDate string) (time.Time, error) {
	for _, dateFormat := range dateFormats {
		if t, err := time.Parse(dateFormat, inDate); err == nil {
			return t, err
		}
	}
	return time.Time{}, fmt.Errorf("Could not parse time string: %s", inDate)
}

func getNCTime(sdsName string, hSubdataset C.GDALDatasetH) ([]string, error) {
	times := []string{}
	metadata := C.GDALGetMetadata(hSubdataset, nil)
	if C.GoString(C.CSLFetchNameValue(metadata, CtimeUnits)) == "" {
		return nil, nil
	}

	timeUnits := C.GoString(C.CSLFetchNameValue(metadata, CtimeUnits))
	timeUnitsSlice := strings.Split(timeUnits, "since")
	stepUnit := durationUnits[strings.Trim(timeUnitsSlice[0], " ")]
	startDate, err := getDate(strings.Trim(timeUnitsSlice[1], " "))
	if err != nil {
		return times, err
	}

	value := C.CSLFetchNameValue(metadata, CncDimTimeValues)
	if value != nil {

		timeStr := C.GoString(value)
		for _, tStr := range strings.Split(strings.Trim(timeStr, "{}"), ",") {
			tF, err := strconv.ParseFloat(tStr, 64)
			if err != nil {
				return times, fmt.Errorf("Problem parsing dates with dataset %s", sdsName)
			}
			secs, _ := math.Modf(tF)
			t := startDate.Add(time.Duration(secs) * stepUnit)
			times = append(times, t.Format("2006-01-02T15:04:05Z"))
		}
		return times, nil
	}
	return times, fmt.Errorf("Dataset %s doesn't contain times", sdsName)
}

func main() {
	C.GDALAllRegister()

	port := flag.Int("p", 1234, "Port RPC.")
	flag.Parse()

	//Creating an instance of struct which implement Arith interface
	ginfo := new(GDALInfo)

	// Register a new rpc server (In most cases, you will use default server only)
	// And register struct we created above by name "Arith"
	// The wrapper method here ensures that only structs which implement Arith interface
	// are allowed to register themselves.
	server := rpc.NewServer()
	server.Register(ginfo)

	// Listen for incoming tcp packets on specified port.
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if e != nil {
		fmt.Printf("listen error:%v\n", e)
	}

	// This statement links rpc server to the socket, and allows rpc server to accept
	// rpc request coming from that socket.
	server.Accept(l)
}
