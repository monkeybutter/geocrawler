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

var AnteMeridianWkt = "POLYGON ((0 90, 0 80, 0 70, 0 60, 0 50, 0 40, 0 30, 0 20, 0 10, 0 0, 0 -10, 0 -20, 0 -30, 0 -40, 0 -50, 0 -60, 0 -70, 0 -80, 0 -90, 180 -80, 180 -70, 180 -60, 180 -50, 180 -40, 180 -30, 180 -20, 180 -10, 180 0,180 10, 180 20, 180 30, 180 40, 180 50, 180 60, 180 70, 180 80, 0 90))"
var PostMeridianWkt = "POLYGON ((-180 -90, -180 -80, -180 -70, -180 -60, -180 -50, -180 -40, -180 -30, -180 -20, -180 -10, -180 0, -180 10, -180 20, -180 30, -180 40, -180 50, -180 60, -180 70, -180 80, -180 90, 0 80, 0 70, 0 60, 0 50, 0 40, 0 30, 0 20, 0 10, 0 0, 0 -10, 0 -20, 0 -30, 0 -40, 0 -50, 0 -60, 0 -70, 0 -80, -180 -90))"
var WGS84WKT = `GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4326"]]`

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

func (p GDALPolygon) Reproject(ProjWKT string) GDALPolygon {
	desSRS := C.OSRNewSpatialReference(C.CString(ProjWKT))

	pszWKT := C.CString("")
	defer C.free(unsafe.Pointer(pszWKT))

	C.OSRExportToWkt(desSRS, &pszWKT)

	//make copy
	newPoly := GDALPolygon{C.OGR_G_Clone(p.Handler)}

	C.OGR_G_TransformTo(newPoly.Handler, desSRS)

	return newPoly
}

func (p GDALPolygon) ReprojectToWGS84() GDALPolygon {
	desSRS := C.OSRNewSpatialReference(C.CString(C.SRS_WKT_WGS84))

	pszWKT := C.CString("")
	defer C.free(unsafe.Pointer(pszWKT))

	C.OSRExportToWkt(desSRS, &pszWKT)

	//make copy
	newPoly := GDALPolygon{C.OGR_G_Clone(p.Handler)}

	err := C.OGR_G_TransformTo(newPoly.Handler, desSRS)

	// Solves Himawari8 problem
	if err == 6 {
		ringCoords := p.AsArray()
		poly := PolygonFromCorners(ringCoords[0][0]-(ringCoords[0][0]*.01), 1, ringCoords[3][0]-(ringCoords[3][0]*.01), -1, p.ProjWKT())
		polyWGS84 := poly.ReprojectToWGS84()
		ringWGS84Coords := polyWGS84.AsArray()
		newPoly = PolygonFromCorners(ringWGS84Coords[0][0], 90, ringWGS84Coords[3][0], -90, polyWGS84.ProjWKT())
	}

	return newPoly
}

func (p GDALPolygon) Intersection(a GDALPolygon) GDALPolygon {
	hGeom := C.OGR_G_Intersection(p.Handler, a.Handler)
	return GDALPolygon{hGeom}
}

func (p GDALPolygon) AsArray() [][]float64 {
	var poly geo.Polygon
	err := poly.UnmarshalWKT(p.ToWKT())

	if err != nil {
		fmt.Println(err)
	}

	return poly.AsArray()[0]
}

func GetPolygon(projWKT, polyWKT string) GDALPolygon {
	ppszData := C.CString(polyWKT)
	defer C.free(unsafe.Pointer(ppszData))

	hSRS := C.OSRNewSpatialReference(nil)
	cProjWKT := C.CString(projWKT)
	defer C.free(unsafe.Pointer(cProjWKT))

	C.OSRImportFromWkt(hSRS, &cProjWKT)

	var hPt C.OGRGeometryH

	_ = C.OGR_G_CreateFromWkt(&ppszData, hSRS, &hPt)

	C.OGR_G_AssignSpatialReference(hPt, hSRS)
	return GDALPolygon{hPt}
}

func GetPolygonFromGeotransform(projWKT string, geoTrans []float64, xSize, ySize int) GDALPolygon {
	var ulX, ulY, lrX, lrY float64
	C.GDALApplyGeoTransform((*C.double)(&geoTrans[0]), C.double(0), C.double(0), (*C.double)(&ulX), (*C.double)(&ulY))
	C.GDALApplyGeoTransform((*C.double)(&geoTrans[0]), C.double(xSize), C.double(ySize), (*C.double)(&lrX), (*C.double)(&lrY))

	polyWKT := fmt.Sprintf("POLYGON ((%f %f,%f %f,%f %f,%f %f,%f %f))", ulX, ulY, ulX, lrY, lrX, lrY, lrX, ulY, ulX, ulY)
	
	return GetPolygon(projWKT, polyWKT)
}

func SplitDateLine(p GDALPolygon) []string {
	postMer := GetPolygon(WGS84WKT, PostMeridianWkt)
	anteMer := GetPolygon(WGS84WKT, AnteMeridianWkt)

	nativePostMer := postMer.Transform(p.ProjWKT())
	nativeAnteMer := anteMer.Transform(p.ProjWKT())
	
	postInters := nativePostMer.Intersection(p)
	result :=  []string{}
	if postInters.Handler != nil {
		result = append(result, postInters.ToWKT())	
	}
	if anteInters.Handler != nil {
		result = append(result, anteInters.ToWKT())	
	}

	return result 
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
