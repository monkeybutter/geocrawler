package main

// #include <stdio.h>
// #include <stdlib.h>
// #include "gdal.h"
// #include "ogr_srs_api.h" /* for SRS calls */
// #include "cpl_string.h"
// #cgo LDFLAGS: -lgdal
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
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	pb "./gdalservice"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
)

var parserStrings map[string]string = map[string]string{"landsat": `LC(?P<mission>\d)(?P<path>\d\d\d)(?P<row>\d\d\d)(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<processing_level>[a-zA-Z0-9]+)_(?P<band>[a-zA-Z0-9]+)`,
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

var parsers map[string]*regexp.Regexp = map[string]*regexp.Regexp{}

var dateFormats []string = []string{"2006-01-02 15:04:05.0", "2006-1-2 15:4:5"}
var durationUnits map[string]time.Duration = map[string]time.Duration{"seconds": time.Second, "hours": time.Hour, "days": time.Hour * 24}
var CWGS84WKT *C.char = C.CString(`GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],TOWGS84[0,0,0,0,0,0,0],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9108"]],AUTHORITY["EPSG","4326"]]","proj4":"+proj=longlat +ellps=WGS84 +towgs84=0,0,0,0,0,0,0 +no_defs `)
var CsubDS *C.char = C.CString("SUBDATASETS")
var CtimeUnits *C.char = C.CString("time#units")
var CncDimTimeValues *C.char = C.CString("NETCDF_DIM_time_VALUES")

var GDALTypes map[C.GDALDataType]string = map[C.GDALDataType]string{0: "Unkown", 1: "Byte", 2: "Uint16", 3: "Int16",
	4: "UInt32", 5: "Int32", 6: "Float32", 7: "Float64",
	8: "CInt16", 9: "CInt32", 10: "CFloat32", 11: "CFloat64",
	12: "TypeCount"}

func init() {
	for key, value := range parserStrings {
		parsers[key] = regexp.MustCompile(value)
	}
}

func extractGDALInfo(in *pb.GeoRequest) (*pb.GeoFile, error) {

	cPath := C.CString(in.FilePath)
	defer C.free(unsafe.Pointer(cPath))
	hDataset := C.GDALOpenShared(cPath, C.GA_ReadOnly)
	hMajorObj := C.GDALMajorObjectH(hDataset)
	defer C.GDALClose(hDataset)

	hDriver := C.GDALGetDatasetDriver(hDataset)
	cShortName := C.GDALGetDriverShortName(hDriver)
	shortName := C.GoString(cShortName)

	metadata := C.GDALGetMetadata(hMajorObj, CsubDS)
	nsubds := C.CSLCount(metadata) / C.int(2)

	var datasets = []*pb.GeoMetaData{}
	if nsubds == C.int(0) {
		// There are no subdatasets
		dsInfo, err := getDataSetInfo(in.FilePath, cPath, shortName)
		if err != nil {
			return &pb.GeoFile{}, err
		}
		datasets = append(datasets, dsInfo)

	} else {
		// There are subdatasets
		for i := C.int(1); i <= nsubds; i++ {
			subDSId := C.CString(fmt.Sprintf("SUBDATASET_%d_NAME", i))
			pszSubdatasetName := C.CSLFetchNameValue(metadata, subDSId)
			dsInfo, err := getDataSetInfo(in.FilePath, pszSubdatasetName, shortName)
			if err != nil {
				return &pb.GeoFile{}, err
			}
			datasets = append(datasets, dsInfo)
		}
	}

	return &pb.GeoFile{FileName: in.FilePath, Driver: shortName, DataSets: datasets}, nil
}

func getDataSetInfo(filename string, dsName *C.char, driverName string) (*pb.GeoMetaData, error) {
	datasetName := C.GoString(dsName)
	hSubdataset := C.GDALOpen(dsName, C.GDAL_OF_READONLY)
	if hSubdataset == nil {
		return &pb.GeoMetaData{}, fmt.Errorf("GDAL could not open dataset: %s", datasetName)
	}
	defer C.GDALClose(hSubdataset)

	hMajorObj := C.GDALMajorObjectH(hSubdataset)

	nameFields, timeStamp := parseName(filename)

	var ncTimes []string
	var err error
	var times []*google_protobuf.Timestamp
	if driverName == "netCDF" {
		ncTimes, err = getNCTime(datasetName, hMajorObj)
	}

	if err == nil && ncTimes != nil {
		for _, timestr := range ncTimes {
			t, err := time.ParseInLocation("2006-01-02T15:04:05Z", timestr, time.UTC)
			if err != nil {
				log.Println(err)
				continue
			}
			tp, err := ptypes.TimestampProto(t)
			if err != nil {
				log.Println(err)
				continue
			}
			times = append(times, tp)
		}
	} else {
		tp, err := ptypes.TimestampProto(timeStamp)
		if err != nil {
			log.Println(err)
		}
		times = append(times, tp)
	}

	hBand := C.GDALGetRasterBand(hSubdataset, 1)
	nOvr := C.GDALGetOverviewCount(hBand)
	ovrs := make([]*pb.Overview, int(nOvr))
	for i := C.int(0); i < nOvr; i++ {
		hOvr := C.GDALGetOverview(hBand, i)
		ovrs[int(i)] = &pb.Overview{XSize: int32(C.GDALGetRasterBandXSize(hOvr)), YSize: int32(C.GDALGetRasterBandYSize(hOvr))}
	}

	projWkt := C.GoString(C.GDALGetProjectionRef(hSubdataset))
	if projWkt == "" {
		projWkt = C.GoString(CWGS84WKT)
	}

	nspace, err := extractNamespace(datasetName)
	if err != nil {
		nspace = nameFields["namespace"]
	}

	dArr := [6]C.double{}
	C.GDALGetGeoTransform(hSubdataset, &dArr[0])
	geot := (*[6]float64)(unsafe.Pointer(&dArr))[:]
	polyWkt := getGeometryWKT(geot, int(C.GDALGetRasterXSize(hSubdataset)), int(C.GDALGetRasterYSize(hSubdataset)))
	cProjWKT := C.CString(projWkt)
	cProj4 := C.getProj4(cProjWKT)
	C.free(unsafe.Pointer(cProjWKT))
	proj4 := C.GoString(cProj4)
	C.free(unsafe.Pointer(cProj4))

	return &pb.GeoMetaData{DatasetName: datasetName, NameSpace: nspace, Type: GDALTypes[C.GDALGetRasterDataType(hBand)],
		RasterCount: int32(C.GDALGetRasterCount(hSubdataset)), TimeStamps: times,
		XSize: int32(C.GDALGetRasterXSize(hSubdataset)), YSize: int32(C.GDALGetRasterYSize(hSubdataset)),
		Polygon: polyWkt, ProjWKT: projWkt, Proj4: proj4, GeoTransform: geot, Overviews: ovrs}, nil

}

func getGeometryWKT(geot []float64, xSize, ySize int) string {
	var ulX, ulY, lrX, lrY C.double
	C.GDALApplyGeoTransform((*C.double)(unsafe.Pointer(&geot[0])), 0, 0, &ulX, &ulY)
	C.GDALApplyGeoTransform((*C.double)(unsafe.Pointer(&geot[0])), C.double(xSize), C.double(ySize), &lrX, &lrY)
	return fmt.Sprintf("POLYGON ((%f %f,%f %f,%f %f,%f %f,%f %f))", ulX, ulY, ulX, lrY, lrX, lrY, lrX, ulY, ulX, ulY)
}

func parseName(filePath string) (map[string]string, time.Time) {
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

/*
	func goStrings(argc C.int, argv **C.char) []string {

			length := int(argc)
				tmpslice := (*[1 << 30]*C.char)(unsafe.Pointer(argv))[:length:length]
					gostrings := make([]string, length)
						for i, s := range tmpslice {
									gostrings[i] = C.GoString(s)
										}
											return gostrings
										}
*/

func getDate(inDate string) (time.Time, error) {
	for _, dateFormat := range dateFormats {
		if t, err := time.Parse(dateFormat, inDate); err == nil {
			return t, err
		}
	}
	return time.Time{}, fmt.Errorf("Could not parse time string: %s", inDate)
}

func getNCTime(sdsName string, hSubdataset C.GDALMajorObjectH) ([]string, error) {
	times := []string{}
	metadata := C.GDALGetMetadata(hSubdataset, nil)
	timeUnits := C.GoString(C.CSLFetchNameValue(metadata, CtimeUnits))
	timeUnitsWords := strings.Split(timeUnits, " ")
	if len(timeUnitsWords) != 4 && timeUnitsWords[1] != "since" {
		return nil, fmt.Errorf("Cannot parse Units string")
	}

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

func extractNamespace(dsName string) (string, error) {
	parts := strings.Split(dsName, ":")
	if len(parts) > 2 {
		return parts[len(parts)-1], nil
	}
	return "", fmt.Errorf("%s does not contain a namespace", dsName)
}

func sendOutput(out *pb.GeoFileAndError, conn net.Conn) error {
	outb, err := proto.Marshal(out)
	if err != nil {
		return err
	}

	_, err = conn.Write(outb)
	if err != nil {
		return err
	}

	return nil

}

func gdalExtractor(conn net.Conn) {
	defer conn.Close()
	out := &pb.GeoFileAndError{GeoFile: &pb.GeoFile{}, Error: "OK"}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, conn)
	if err != nil {
		out.Error = fmt.Sprintf("Error reading data %d from socket: %v", n, err)
		sendOutput(out, conn)
	}

	in := new(pb.GeoRequest)
	err = proto.Unmarshal(buf.Bytes(), in)
	if err != nil {
		out.Error = fmt.Sprintf("Error unmarshaling protobuf request: %v", err)
		sendOutput(out, conn)
	}

	gInfo, err := extractGDALInfo(in)
	if err != nil {
		out.Error = fmt.Sprintf("Error warping raster: %v", err)
		sendOutput(out, conn)
	}
	out.GeoFile = gInfo

	_ = sendOutput(out, conn)
	// TODO "How do we report this error?"

}

func main() {
	C.GDALAllRegister()

	addr := os.Args[1]
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: addr, Net: "unix"})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer os.Remove(addr)

	fmt.Printf("Listening in %s\n", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
			return
		}
		gdalExtractor(conn)
	}
}
