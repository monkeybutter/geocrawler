package geolib

// #include "gdal.h"
// #include "ogr_srs_api.h" /* for SRS calls */
// #include "cpl_string.h"
// #cgo LDFLAGS: -lgdal
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

type GDALDataSet struct {
	DataSetName  string    `json:"ds_name"`
	RasterCount  int       `json:"raster_count"`
	Type         string    `json:"array_type"`
	XSize        int       `json:"x_size"`
	YSize        int       `json:"y_size"`
	ProjWKT        string    `json:"proj_wkt"`
	GeoTransform []float64 `json:"geotransform"`
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

	// To hold the SRS
	hSRS := C.OSRNewSpatialReference(nil)

	// To hold the geotransform
	dArr := [6]C.double{}

	hSubdataset := C.GDALOpenEx(dsName, C.GDAL_OF_READONLY|C.GDAL_OF_RASTER, nil, nil, nil)
	if hSubdataset == nil {
		return GDALDataSet{}
	}
	defer C.GDALClose(hSubdataset)

	hBand := C.GDALGetRasterBand(hSubdataset, 1)

	pszProjection := C.GDALGetProjectionRef(hSubdataset)
	C.OSRImportFromWkt(hSRS, &pszProjection)
	
	pszWkt := C.GDALGetProjectionRef(hSubdataset)

	C.GDALGetGeoTransform(hSubdataset, &dArr[0])

	fArr := (*[6]float64)(unsafe.Pointer(&dArr))

	return GDALDataSet{C.GoString(dsName), int(C.GDALGetRasterCount(hSubdataset)), GDALTypes[C.GDALGetRasterDataType(hBand)],
			int(C.GDALGetRasterXSize(hSubdataset)), int(C.GDALGetRasterYSize(hSubdataset)), C.GoString(pszWkt), fArr[:]}

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
			pszSubdatasetName := C.CPLStrdup(C.CSLFetchNameValue(metadata, subDSId))
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
