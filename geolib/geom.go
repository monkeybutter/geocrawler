package geolib

// #include "gdal.h"
// #include "ogr_api.h"
// #include "ogr_srs_api.h"
// #cgo LDFLAGS: -lgdal
import "C"

import (
	geo "bitbucket.org/monkeyforecaster/geometry"
	"fmt"
	"unsafe"
)

/*
type Geometry struct {
	Type string      `json:"type"`
	Coordinates [][][]float32 `json:"coordinates"`
}

type GeoJSON struct {
	Type string      `json:"type"`
	Geometry Geometry `json:"geometry"`
	Properties map[string]string `json:"properties"`
}
*/

type GDALPolygon struct {
	Handler C.OGRGeometryH
}

/*
func (p GDALPolygon) ToGeoJSON() string {
	return C.GoString(C.OGR_G_ExportToJson(p.Handler))
}
*/

func (p GDALPolygon) ToWKT() string {
	ppszSrcText := C.CString("")
	defer C.free(unsafe.Pointer(ppszSrcText))

	C.OGR_G_ExportToWkt(p.Handler, &ppszSrcText)
	return C.GoString(ppszSrcText)

}

func (p GDALPolygon) Proj4() string {
	pszProj4 := C.CString("")
	defer C.free(unsafe.Pointer(pszProj4))

	C.OSRExportToProj4(C.OGR_G_GetSpatialReference(p.Handler), &pszProj4)
	return C.GoString(pszProj4)
}

func (p GDALPolygon) ProjWKT() string {
	pszProjWKT := C.CString("")
	defer C.free(unsafe.Pointer(pszProjWKT))

	C.OSRExportToWkt(C.OGR_G_GetSpatialReference(p.Handler), &pszProjWKT)
	return C.GoString(pszProjWKT)
}

func (p GDALPolygon) ReprojectToWGS84() GDALPolygon {
	desSRS := C.OSRNewSpatialReference(C.CString(C.SRS_WKT_WGS84))

	pszWKT := C.CString("")
	defer C.free(unsafe.Pointer(pszWKT))

	C.OSRExportToWkt(desSRS, &pszWKT)

	//make copy
	newPoly := GDALPolygon{C.OGR_G_Clone(p.Handler)}

	err := C.OGR_G_TransformTo(newPoly.Handler, desSRS)
	if err == 6 {
		ringCoords := p.AsArray()
		poly := PolygonFromCorners(ringCoords[0][0]-(ringCoords[0][0]*.01), 1, ringCoords[3][0]-(ringCoords[3][0]*.01), -1, p.ProjWKT())
		polyWGS84 := poly.ReprojectToWGS84()
		ringWGS84Coords := polyWGS84.AsArray()
		newPoly = PolygonFromCorners(ringWGS84Coords[0][0], 90, ringWGS84Coords[3][0], -90, polyWGS84.ProjWKT())
	}

	return newPoly
}

func (p GDALPolygon) AsArray() [][]float64 {
	var poly geo.Polygon
	err := poly.UnmarshalWKT(p.ToWKT())

	if err != nil {
		fmt.Println(err)
	}

	return poly.AsArray()[0]
}

func GetPolygon(projWKT string, geoTrans []float64, xSize, ySize int) GDALPolygon {

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
	p := GDALPolygon{hPt}

	return p
}

func SplitDateLine(p GDALPolygon) []string {
	return []string{p.ToWKT()}
}

func PolygonFromCorners(ulX, ulY, lrX, lrY float64, projWKT string) GDALPolygon {

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
	p := GDALPolygon{hPt}

	return p
}
