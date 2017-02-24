package processor

// #include "gdal.h"
// #include "ogr_api.h"
// #include "ogr_srs_api.h"
// #cgo LDFLAGS: -lgdal
import "C"

import (
	"../rpcflow"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type GeoMetaData struct {
	DataSetName    string            `json:"ds_name"`
	NameSpace      string            `json:"namespace"`
	TimeStamps     []time.Time       `json:"timestamps"`
	FileNameFields map[string]string `json:"filename_fields"`
	Polygon        string            `json:"polygon"`
	RasterCount    int               `json:"raster_count"`
	Type           string            `json:"array_type"`
	XSize          int               `json:"x_size"`
	YSize          int               `json:"y_size"`
	Overviews      []rpcflow.Overview `json:"overviews"`
	ProjWKT        string            `json:"proj_wkt"`
	Proj4          string            `json:"proj4"`
	GeoTransform   []float64         `json:"geotransform"`
}

type GeoFile struct {
	FileName string        `json:"filename,omitempty"`
	Driver   string        `json:"file_type"`
	DataSets []GeoMetaData `json:"geo_metadata"`
}

var parserStrings map[string]string = map[string]string{"landsat": `L(?P<sensor>[COT])(?P<mission>\d)(?P<path>\d\d\d)(?P<row>\d\d\d)(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<processing_level>[a-zA-Z0-9]+)_(?P<namespace>[a-zA-Z0-9]+)`,
	"modis43A4":     `^MCD43A4.A(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d).(?P<horizontal>h\d\d)(?P<vertical>v\d\d).(?P<resolution>\d\d\d).[0-9]+`,
	"modis1":        `^(?P<product>MCD\d\d[A-Z]\d).A(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d).(?P<horizontal>h\d\d)(?P<vertical>v\d\d).(?P<resolution>\d\d\d).[0-9]+`,
	"modis2":        `M(?P<satellite>[OD|YD])(?P<product>[0-9]+_[A-Z0-9]+).A[0-9]+.[0-9]+.(?P<collection_version>\d\d\d).(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)`,
	"modisJP":       `^(?P<product>FC).v302.(?P<root_product>MCD\d\d[A-Z]\d).h(?P<horizontal>\d\d)v(?P<vertical>\d\d).(?P<year>\d\d\d\d).(?P<resolution>\d\d\d).`,
	"modisJP_LR":    `^(?P<product>FC_LR).v302.(?P<root_product>MCD\d\d[A-Z]\d).h(?P<horizontal>\d\d)v(?P<vertical>\d\d).(?P<year>\d\d\d\d).(?P<resolution>\d\d\d).`,
	"himawari8":     `^(?P<year>\d\d\d\d)(?P<month>\d\d)(?P<day>\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)-P1S-(?P<product>ABOM[0-9A-Z_]+)-PRJ_GEOS141_(?P<resolution>\d+)-HIMAWARI8-AHI`,
	"agdc_landsat1": `LS(?P<mission>\d)_(?P<sensor>[A-Z]+)_(?P<correction>[A-Z]+)_(?P<epsg>\d+)_(?P<x_coord>-?\d+)_(?P<y_coord>-?\d+)_(?P<year>\d\d\d\d).`,
	"elevation_ga":  `^Elevation_1secSRTM_DEMs_v1.0_DEM-S_Tiles_e(?P<longitude>\d+)s(?P<latitude>\d+)dems.nc$`,
	"chirps2.0":     `^chirps-v2.0.(?P<year>\d\d\d\d).dekads.nc$`,
	"era-interim":   `^(?P<product>[a-z0-9]+)_3hrs_ERAI_historical_fc-sfc_(?P<start_year>\d\d\d\d)(?P<start_month>\d\d)(?P<start_day>\d\d)_(?P<end_year>\d\d\d\d)(?P<end_month>\d\d)(?P<end_day>\d\d).nc$`,
	"agdc_landsat2": `LS(?P<mission>\d)_OLI_(?P<sensor>[A-Z]+)_(?P<product>[A-Z]+)_(?P<epsg>\d+)_(?P<x_coord>-?\d+)_(?P<y_coord>-?\d+)_(?P<year>\d\d\d\d).`,
	"agdc_dem":      `SRTM_(?P<product>[A-Z]+)_(?P<x_coord>-?\d+)_(?P<y_coord>-?\d+)_(?P<year>\d\d\d\d)(?P<month>\d\d)(?P<day>\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)`}


func getPolygonFromGeoTransform(projWKT string, geoTrans []float64, xSize, ySize int) C.OGRGeometryH {
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
	C.OGR_G_CreateFromWkt(&ppszData, hSRS, &hPt)

	return C.OGR_G_Clone(hPt)
}

func Geom2WKT(geom C.OGRGeometryH) string {
	var ppszSrcText *C.char
	defer C.free(unsafe.Pointer(ppszSrcText))

	C.OGR_G_ExportToWkt(geom, &ppszSrcText)
	return C.GoString(ppszSrcText)
}

func Proj4(geom C.OGRGeometryH) string {
	var pszProj4 *C.char
	defer C.free(unsafe.Pointer(pszProj4))

	C.OSRExportToProj4(C.OGR_G_GetSpatialReference(geom), &pszProj4)
	return C.GoString(pszProj4)
}

func ProjWKT(geom C.OGRGeometryH) string {
	var pszProjWKT *C.char
	defer C.free(unsafe.Pointer(pszProjWKT))

	C.OSRExportToWkt(C.OGR_G_GetSpatialReference(geom), &pszProjWKT)
	return C.GoString(pszProjWKT)
}

type GeoParser struct {
	In      chan rpcflow.GDALFile
	Out     chan GeoFile
	Error   chan error
	parsers map[string]*regexp.Regexp
}

func NewGeoParser(errChan chan error) *GeoParser {
	prs := map[string]*regexp.Regexp{}
	for key, value := range parserStrings {
		prs[key] = regexp.MustCompile(value)
	}

	return &GeoParser{
		In:      make(chan rpcflow.GDALFile, 100),
		Out:     make(chan GeoFile, 100),
		Error:   errChan,
		parsers: prs,
	}
}

func (gt *GeoParser) Run() {
	defer close(gt.Out)

	for gdalFile := range gt.In {
		geoFile := GeoFile{FileName: gdalFile.FileName, Driver: gdalFile.Driver}
		nameFields, timeStamp := parseName(gdalFile.FileName, gt.parsers)
		/*if nameFields == nil {
			gt.Error <- fmt.Errorf("GDALFile %v non parseable", gdalFile)
			//continue
		}*/
		for _, ds := range gdalFile.DataSets {
			if ds.ProjWKT != "" {
				poly := getPolygonFromGeoTransform(ds.ProjWKT, ds.GeoTransform, ds.XSize, ds.YSize)
				hSRS := C.OGR_G_GetSpatialReference(poly)
				defer C.OSRDestroySpatialReference(hSRS)

				var times []time.Time
				if nc_times, ok := ds.Extras["nc_times"]; ok {
					for _, timestr := range nc_times {
						t, err := time.ParseInLocation("2006-01-02T15:04:05Z", timestr, time.UTC)
						if err != nil {
							gt.Error <- err
							continue
						}
						times = append(times, t)
					}
				} else {
					times = []time.Time{timeStamp}
				}
				nspace, err := extractNamespace(ds.DataSetName)
				if err != nil {
					nspace = nameFields["namespace"]
				}

				geoFile.DataSets = append(geoFile.DataSets, GeoMetaData{DataSetName: ds.DataSetName, NameSpace: nspace, TimeStamps: times, FileNameFields: nameFields, Polygon: Geom2WKT(poly), RasterCount: ds.RasterCount, Type: ds.Type, XSize: ds.XSize, YSize: ds.YSize, Overviews: ds.Overviews, ProjWKT: ProjWKT(poly), Proj4: Proj4(poly), GeoTransform: ds.GeoTransform})
			}
		}

		gt.Out <- geoFile
	}
}

func extractNamespace(dsName string) (string, error) {
	parts := strings.Split(dsName, ":")
	if len(parts) > 2 {
		return parts[len(parts)-1], nil
	}
	return "", fmt.Errorf("%s does not contain a namespace")
}

func parseName(filePath string, parsers map[string]*regexp.Regexp) (map[string]string, time.Time) {

	for _, r := range parsers {
		_, fileName := filepath.Split(filePath)

		if r.MatchString(fileName) {
			match := r.FindStringSubmatch(fileName)

			result := make(map[string]string)
			for i, name := range r.SubexpNames() {
				if i != 0 {
					result[name] = match[i]
				}
			}

			return result, parseTime(result)
		}
	}

	return nil, time.Time{}
}

func parseTime(nameFields map[string]string) time.Time {
	if _, ok := nameFields["year"]; ok {
		year, _ := strconv.Atoi(nameFields["year"])
		t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		if _, ok := nameFields["julian_day"]; ok {
			julianDay, _ := strconv.Atoi(nameFields["julian_day"])
			t = t.Add(time.Hour * 24 * time.Duration(julianDay-1))
		}
		if _, ok := nameFields["month"]; ok {
			if _, ok := nameFields["day"]; ok {
				month, _ := strconv.Atoi(nameFields["month"])
				day, _ := strconv.Atoi(nameFields["day"])
				t = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			}
		}
		if _, ok := nameFields["hour"]; ok {
			hour, _ := strconv.Atoi(nameFields["hour"])
			t = t.Add(time.Hour * time.Duration(hour))
		}
		if _, ok := nameFields["minute"]; ok {
			minute, _ := strconv.Atoi(nameFields["minute"])
			t = t.Add(time.Minute * time.Duration(minute))
		}
		if _, ok := nameFields["second"]; ok {
			second, _ := strconv.Atoi(nameFields["second"])
			t = t.Add(time.Second * time.Duration(second))
		}

		return t
	}
	return time.Time{}
}
