package geolib

// #include "gdal.h"
// #include "ogr_srs_api.h" /* for SRS calls */
// #include "cpl_string.h"
// #cgo LDFLAGS: -lgdal
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const WGS84WKT = `GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],TOWGS84[0,0,0,0,0,0,0],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9108"]],AUTHORITY["EPSG","4326"]]","proj4":"+proj=longlat +ellps=WGS84 +towgs84=0,0,0,0,0,0,0 +no_defs `
var CWGS84WKT *C.char = C.CString(`GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],TOWGS84[0,0,0,0,0,0,0],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9108"]],AUTHORITY["EPSG","4326"]]","proj4":"+proj=longlat +ellps=WGS84 +towgs84=0,0,0,0,0,0,0 +no_defs `)

type GDALDataSet struct {
	DataSetName  string    `json:"ds_name"`
	RasterCount  int       `json:"raster_count"`
	Type         string    `json:"array_type"`
	XSize        int       `json:"x_size"`
	YSize        int       `json:"y_size"`
	ProjWKT      string    `json:"proj_wkt"`
	GeoTransform []float64 `json:"geotransform"`
	Extras map[string][]string `json:"extra_metadata"`
}

type GDALFile struct {
	Driver   string        `json:"file_type"`
	DataSets []GDALDataSet `json:"datasets"`
}

var GDALTypes map[C.GDALDataType]string = map[C.GDALDataType]string{0: "Unkown", 1: "Byte", 2: "Uint16", 3: "Int16",
	4: "UInt32", 5: "Int32", 6: "Float32", 7: "Float64",
	8: "CInt16", 9: "CInt32", 10: "CFloat32", 11: "CFloat64",
	12: "TypeCount"}

func GetDataSetInfo(dsName *C.char) GDALDataSet {
	datasetName := C.GoString(dsName)
	hSubdataset := C.GDALOpenEx(dsName, C.GDAL_OF_READONLY|C.GDAL_OF_RASTER, nil, nil, nil)
	if hSubdataset == nil {
		return GDALDataSet{}
	}
	defer C.GDALClose(hSubdataset)

	extras := map[string][]string{}	
	nc_times, err := GetNCTime(datasetName, hSubdataset)
	if err == nil {
		extras["nc_times"] = nc_times
	}

	hBand := C.GDALGetRasterBand(hSubdataset, 1)
	pszWkt := C.GDALGetProjectionRef(hSubdataset)
	if C.GoString(pszWkt) == "" {
		pszWkt = CWGS84WKT
	}

	// To hold the geotransform
	dArr := [6]C.double{}
	C.GDALGetGeoTransform(hSubdataset, &dArr[0])
	fArr := (*[6]float64)(unsafe.Pointer(&dArr))

	return GDALDataSet{datasetName, int(C.GDALGetRasterCount(hSubdataset)), GDALTypes[C.GDALGetRasterDataType(hBand)],
		int(C.GDALGetRasterXSize(hSubdataset)), int(C.GDALGetRasterYSize(hSubdataset)), C.GoString(pszWkt), fArr[:], extras}

}

func GetNCTime(sdsName string, hSubdataset C.GDALDatasetH) ([]string, error) {
	nameParts := strings.Split(sdsName, ":")
	times := []string{}
	
	if len(nameParts) == 3 {
		if nameParts[0] == "NETCDF" {
			metadata := C.GDALGetMetadata(hSubdataset, C.CString(""))
			value := C.CSLFetchNameValue(metadata, C.CString("NETCDF_DIM_time_VALUES"))
			if value != nil {

				timeStr := C.GoString(value)
				for _, tStr := range strings.Split(strings.Trim(timeStr, "{}"), ",") {
					tF, err := strconv.ParseFloat(tStr, 64)
					if err != nil {
						return times, errors.New("Problem parsing dates")
					}
					secs, _ := math.Modf(tF)
					t := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
					t = t.Add(time.Second * time.Duration(secs))
					times = append(times, t.Format("2006-01-02T15:04:05Z"))
				}
				return times, nil
			}
		}
	}
	return times, errors.New("Dataset doesn't contain times")
}

func GetGDALMetadata(path string) ([]byte, error) {

	// Register all GDAL Drivers
	C.GDALAllRegister()
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	hDataset := C.GDALOpenEx(cPath, C.GDAL_OF_READONLY|C.GDAL_OF_RASTER, nil, nil, nil)
	if hDataset == nil {
		return nil, nil
	}
	defer C.GDALClose(hDataset)

	metadata := C.GDALGetMetadata(hDataset, C.CString("SUBDATASETS"))
	nsubds := C.CSLCount(metadata) / C.int(2)

	var datasets = []GDALDataSet{}
	if nsubds == C.int(0) {
		// There are no subdatasets
		datasets = append(datasets, GetDataSetInfo(cPath))

	} else {
		// There are subdatasets
		for i := C.int(1); i <= nsubds; i++ {
			subDSId := C.CString(fmt.Sprintf("SUBDATASET_%d_NAME", i))
			defer C.free(unsafe.Pointer(subDSId))
			pszSubdatasetName := C.CSLFetchNameValue(metadata, subDSId)
			datasets = append(datasets, GetDataSetInfo(pszSubdatasetName))
		}
	}

	hDriver := C.GDALGetDatasetDriver(hDataset)
	shortName := C.GDALGetDriverShortName(hDriver)

	return json.Marshal(GDALFile{C.GoString(shortName), datasets})
}

func GDALMetadataPrinter(path string, workers *ConcLimiter) {
	defer workers.Decrease()

	json, err := GetGDALMetadata(path)
	if json != nil && err == nil {
		fmt.Printf("%s\t%s\n", path, string(json))
	}
}
