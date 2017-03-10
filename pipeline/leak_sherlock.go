package main

// #include <stdio.h>
// #include <stdlib.h>
// #include "gdal.h"
// #include "ogr_api.h"
// #include "ogr_srs_api.h"
// #include "cpl_string.h"
// #include "cpl_conv.h"
// #cgo LDFLAGS: -lgdal
//
//char *getProj4(char *projWKT)
//{
//	char *pszProj4;
//	char *result;
//	OGRSpatialReferenceH hSRS;
//
//	hSRS = OSRNewSpatialReference(projWKT);
//	OSRExportToProj4(hSRS, &pszProj4);
//	result = strdup(pszProj4);
//	
//	OSRDestroySpatialReference(hSRS);
//	CPLFree(pszProj4);
//	
//	return result; 
//}
import "C"

import (
	"../rpcflow"
	"fmt"
	"unsafe"
)

var CsubDS *C.char = C.CString("SUBDATASETS")
var GDALTypes map[C.GDALDataType]string = map[C.GDALDataType]string{0: "Unkown", 1: "Byte", 2: "Uint16", 3: "Int16",
	4: "UInt32", 5: "Int32", 6: "Float32", 7: "Float64",
	8: "CInt16", 9: "CInt32", 10: "CFloat32", 11: "CFloat64",
	12: "TypeCount"}

func getMetadata(filePath string) rpcflow.GDALFile {
	cPath := C.CString(filePath)
	defer C.free(unsafe.Pointer(cPath))

	hDataset := C.GDALOpenShared(cPath, C.GDAL_OF_READONLY)
	if hDataset == nil {
		fmt.Printf("GDAL could not open dataset: %s", filePath)
		return rpcflow.GDALFile{}
	}
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
			return rpcflow.GDALFile{}
		}
		
		datasets = append(datasets, dsInfo)
	}
	
	return rpcflow.GDALFile{filePath, shortName, datasets}
}

func getDataSetInfo(dsName *C.char, driverName string) (rpcflow.GDALDataSet, error) {

	datasetName := C.GoString(dsName)
	hSubdataset := C.GDALOpenShared(dsName, C.GDAL_OF_READONLY)
	if hSubdataset == nil {
		return rpcflow.GDALDataSet{}, fmt.Errorf("GDAL could not open dataset: %s", datasetName)
	}
	defer C.GDALClose(hSubdataset)
	
	hBand := C.GDALGetRasterBand(hSubdataset, 1)

	extras := map[string][]string{}

	nOvr := C.GDALGetOverviewCount(hBand)
	ovrs := make([]rpcflow.Overview, int(nOvr))
	for i := C.int(0); i < nOvr; i++ {
		hOvr := C.GDALGetOverview(hBand, i)
		ovrs[int(i)] = rpcflow.Overview{int(C.GDALGetRasterBandXSize(hOvr)), int(C.GDALGetRasterBandYSize(hOvr))}
	}

	pszWkt := C.GDALGetProjectionRef(hSubdataset)

	// To hold the geotransform
	dArr := [6]C.double{}
	C.GDALGetGeoTransform(hSubdataset, &dArr[0])
	fArr := (*[6]float64)(unsafe.Pointer(&dArr))

	return rpcflow.GDALDataSet{datasetName, int(C.GDALGetRasterCount(hSubdataset)), GDALTypes[C.GDALGetRasterDataType(hBand)],
		int(C.GDALGetRasterXSize(hSubdataset)), int(C.GDALGetRasterYSize(hSubdataset)), C.GoString(pszWkt), fArr[:], ovrs, extras}, nil
}

func main() {
	C.GDALAllRegister()
	for true {
		fmt.Println(getMetadata("/vsis3/landsat-pds/L8/001/004/LC80010042014272LGN00/LC80010042014272LGN00_B9.TIF"))
	}
}
