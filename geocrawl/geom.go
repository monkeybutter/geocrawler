package geocrawl

// #include "gdal.h"
// #include "ogr_api.h"
// #include "ogr_srs_api.h"
// #cgo LDFLAGS: -lgdal
import "C"

import (
	"fmt"
	"unsafe"
)

type Geometry struct {
	Type string      `json:"type"`
	Coordinates [][][]float32 `json:"coordinates"`
}

type GeoJSON struct {
	Type string      `json:"type"`
	Geometry Geometry `json:"geometry"`
	Properties map[string]string `json:"properties"`
}

type Polygon struct {
	Handler C.OGRGeometryH
}

func (p Polygon) ToGeoJSON() string {
	return C.GoString(C.OGR_G_ExportToJson(p.Handler))
}

func (p Polygon) ToWKT() string {
	ppszSrcText := C.CString("")
	defer C.free(unsafe.Pointer(ppszSrcText))

	C.OGR_G_ExportToWkt(p.Handler, &ppszSrcText)
	return C.GoString(ppszSrcText)

}

func (p Polygon) Proj4() string {
	pszProj4 := C.CString("")
	defer C.free(unsafe.Pointer(pszProj4))

	C.OSRExportToProj4(C.OGR_G_GetSpatialReference(p.Handler), &pszProj4)
	return C.GoString(pszProj4)
}

func (p Polygon) ProjWKT() string {

	pszProjWKT := C.CString("")
	defer C.free(unsafe.Pointer(pszProjWKT))

	C.OSRExportToWkt(C.OGR_G_GetSpatialReference(p.Handler), &pszProjWKT)
	return C.GoString(pszProjWKT)

}

func (p Polygon) ReprojectToWGS84() Polygon {
	desSRS := C.OSRNewSpatialReference(C.CString(C.SRS_WKT_WGS84))

	pszWKT := C.CString("")
	defer C.free(unsafe.Pointer(pszWKT))

	C.OSRExportToWkt(desSRS, &pszWKT)

	//make copy
	newPoly := Polygon{C.OGR_G_Clone(p.Handler)}

	// Get error?
	_ = C.OGR_G_TransformTo(newPoly.Handler, desSRS)

	return newPoly
}

func GetPolygon(projWKT string, geoTrans []float64, xSize, ySize int) Polygon {

	var ulX, ulY, lrX, lrY float64
	C.GDALApplyGeoTransform((*C.double)(&geoTrans[0]), C.double(0), C.double(0), (*C.double)(&ulX), (*C.double)(&ulY))
	C.GDALApplyGeoTransform((*C.double)(&geoTrans[0]), C.double(xSize), C.double(ySize), (*C.double)(&lrX), (*C.double)(&lrY))

	polyWKT := fmt.Sprintf("POLYGON ((%f %f,%f %f,%f %f,%f %f,%f %f))", ulX, ulY, ulX, lrY, lrX, lrY, lrX, ulY, ulX, ulY)

	ppszData := C.CString(polyWKT)
	defer C.free(unsafe.Pointer(ppszData))

	hSRS := C.OSRNewSpatialReference(nil)
	cProjWKT := C.CString(projWKT)
	defer C.free(unsafe.Pointer(cProjWKT))

	C.OSRImportFromWkt(hSRS, &cProjWKT)

	var hPt C.OGRGeometryH

	_ = C.OGR_G_CreateFromWkt(&ppszData, hSRS, &hPt)

	C.OGR_G_AssignSpatialReference(hPt, hSRS)
	p := Polygon{hPt}

	return p
}
