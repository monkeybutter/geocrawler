package processor

// #include "gdal.h"
// #include "ogr_srs_api.h" /* for SRS calls */
// #include "cpl_string.h"
// #cgo LDFLAGS: -lgdal
import "C"

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type GDALDataSet struct {
	DataSetName  string              `json:"ds_name"`
	RasterCount  int                 `json:"raster_count"`
	Type         string              `json:"array_type"`
	XSize        int                 `json:"x_size"`
	YSize        int                 `json:"y_size"`
	ProjWKT      string              `json:"proj_wkt"`
	GeoTransform []float64           `json:"geotransform"`
	Extras       map[string][]string `json:"extra_metadata"`
}

type GDALFile struct {
	FileName  string              `json:"file_name"`
	Driver   string        `json:"file_type"`
	DataSets []GDALDataSet `json:"datasets"`
}

var dateFormats []string = []string{"2006-01-02 15:04:05.0", "2006-1-2 15:4:5"}

var durationUnits map[string]time.Duration = map[string]time.Duration{"seconds": time.Second, "hours": time.Hour, "days": time.Hour * 24}

var CWGS84WKT *C.char = C.CString(`GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],TOWGS84[0,0,0,0,0,0,0],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9108"]],AUTHORITY["EPSG","4326"]]","proj4":"+proj=longlat +ellps=WGS84 +towgs84=0,0,0,0,0,0,0 +no_defs `)

var GDALTypes map[C.GDALDataType]string = map[C.GDALDataType]string{0: "Unkown", 1: "Byte", 2: "Uint16", 3: "Int16",
	4: "UInt32", 5: "Int32", 6: "Float32", 7: "Float64",
	8: "CInt16", 9: "CInt32", 10: "CFloat32", 11: "CFloat64",
	12: "TypeCount"}

type GDALInfo struct {
	ID    int
	In    chan string
	Out   chan GDALFile
	Error chan error
}

func NewGDALInfo(errChan chan error) *GDALInfo {
	return &GDALInfo{
		In:    make(chan string),
		Out:   make(chan GDALFile),
		Error: errChan,
	}
}

func (gi *GDALInfo) Run() {
	defer close(gi.Out)
	C.GDALAllRegister()

	for path := range gi.In {
		cPath := C.CString(path)

		hDataset := C.GDALOpenShared(cPath, C.GDAL_OF_READONLY)
		if hDataset == nil {
			gi.Error <- fmt.Errorf("GDAL could not open dataset: %s", path)
			continue
		}

		hDriver := C.GDALGetDatasetDriver(hDataset)
		shortName := C.GoString(C.GDALGetDriverShortName(hDriver))

		metadata := C.GDALGetMetadata(hDataset, C.CString("SUBDATASETS"))
		nsubds := C.CSLCount(metadata) / C.int(2)

		var datasets = []GDALDataSet{}
		if nsubds == C.int(0) {
			// There are no subdatasets
			dsInfo, err := getDataSetInfo(cPath, shortName)
			if err != nil {
				gi.Error <- err
				return
			}
			datasets = append(datasets, dsInfo)

		} else {
			// There are subdatasets
			for i := C.int(1); i <= nsubds; i++ {
				subDSId := C.CString(fmt.Sprintf("SUBDATASET_%d_NAME", i))
				pszSubdatasetName := C.CSLFetchNameValue(metadata, subDSId)
				C.free(unsafe.Pointer(subDSId))
				dsInfo, err := getDataSetInfo(pszSubdatasetName, shortName)
				if err != nil {
					gi.Error <- err
					return
				}
				datasets = append(datasets, dsInfo)
			}
		}

		C.free(unsafe.Pointer(cPath))
		C.GDALClose(hDataset)

		gi.Out <- GDALFile{path, shortName, datasets}
	}
}

func getDataSetInfo(dsName *C.char, driverName string) (GDALDataSet, error) {
	datasetName := C.GoString(dsName)
	hSubdataset := C.GDALOpenShared(dsName, C.GDAL_OF_READONLY)
	if hSubdataset == nil {
		return GDALDataSet{}, fmt.Errorf("GDAL could not open dataset: %s", C.GoString(dsName))
	}
	defer C.GDALClose(hSubdataset)

	extras := map[string][]string{}
	if driverName == "netCDF" {
		nc_times, err := getNCTime(datasetName, hSubdataset)
		if err == nil {
			extras["nc_times"] = nc_times
		}
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
		int(C.GDALGetRasterXSize(hSubdataset)), int(C.GDALGetRasterYSize(hSubdataset)), C.GoString(pszWkt), fArr[:], extras}, nil

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
	timeUnits := C.GoString(C.CSLFetchNameValue(metadata, C.CString("time#units")))
	timeUnitsSlice := strings.Split(timeUnits, "since")
	stepUnit := durationUnits[strings.Trim(timeUnitsSlice[0], " ")]
	startDate, err := getDate(strings.Trim(timeUnitsSlice[1], " "))
	if err != nil {
		return times, err
	}

	value := C.CSLFetchNameValue(metadata, C.CString("NETCDF_DIM_time_VALUES"))
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
