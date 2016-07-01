package main

import (
	"../geolib"
	geo "bitbucket.org/monkeyforecaster/geometry"
	"bufio"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	_ "gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GeoMetaData struct {
	DataSetName    string            `json:"ds_name" bson:"ds_name"`
	TimeStamps     []time.Time       `json:"timestamps" bson:"timestamps"`
	FileNameFields map[string]string `json:"filename_fields" bson:"filename_fields"`
	Location       geo.Geometry      `json:"location" bson:"location"`
	RasterCount    int               `json:"raster_count" bson:"raster_count"`
	Type           string            `json:"array_type" bson:"array_type"`
	XSize          int               `json:"x_size" bson:"x_size"`
	YSize          int               `json:"y_size" bson:"y_size"`
	ProjWKT        string            `json:"proj_wkt" bson:"proj_wkt"`
	GeoTransform   []float64         `json:"geotransform" bson:"geotransform"`
}

type GeoFile struct {
	FileName string        `json:"filename" bson:"filename"`
	Driver   string        `json:"file_type" bson:"file_type"`
	DataSets []GeoMetaData `json:"geo_metadata" bson:"geo_metadata"`
}

//20160103032000-P1S-ABOM_BRF_B03-PRJ_GEOS141_1000-HIMAWARI8-AHI.nc
//20160107115000-P1S-ABOM_OBS_B07-PRJ_GEOS141_2000-HIMAWARI8-AHI.nc
//2016|03|03|00|20|00-P1S-ABOM_BRF_B01-PRJ_GEOS141_1000-HIMAWARI8-AHI.nc
//SRTM_DEM_9_-49_20000221115400000000.nc
//LS8_OLI_TIRS_PQ_3577_9_-49_20130417000652240357

var parserStrings map[string]string = map[string]string{"landsat": `LC(?P<mission>\d)(?P<path>\d\d\d)(?P<row>\d\d\d)(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<processing_level>[a-zA-Z0-9]+)_(?P<band>[a-zA-Z0-9]+)`,
	"modis1":        `M(?P<satellite>[OD|YD])(?P<product>[0-9]+_[A-Z0-9]+).A[0-9]+.[0-9]+.(?P<collection_version>\d\d\d).(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)`,
	"modis2":        `^(?P<product>MCD\d\d[A-Z]\d).A[0-9]+.(?P<horizontal>h\d\d)(?P<vertical>v\d\d).(?P<resolution>\d\d\d).(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)`,
	"himawari8":     `^(?P<year>\d\d\d\d)(?P<month>\d\d)(?P<day>\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)-P1S-(?P<product>ABOM[0-9A-Z_]+)-PRJ_GEOS141_(?P<resolution>\d+)-HIMAWARI8-AHI`,
	"agdc_landsat1": `LS(?P<mission>\d)_(?P<sensor>[A-Z]+)_(?P<correction>[A-Z]+)_(?P<epsg>\d+)_(?P<x_coord>-?\d+)_(?P<y_coord>-?\d+)_(?P<year>\d\d\d\d).`,
	"agdc_landsat2": `LS(?P<mission>\d)_OLI_(?P<sensor>[A-Z]+)_(?P<product>[A-Z]+)_(?P<epsg>\d+)_(?P<x_coord>-?\d+)_(?P<y_coord>-?\d+)_(?P<year>\d\d\d\d).`,
	"agdc_dem":      `SRTM_(?P<product>[A-Z]+)_(?P<x_coord>-?\d+)_(?P<y_coord>-?\d+)_(?P<year>\d\d\d\d)(?P<month>\d\d)(?P<day>\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)`}

var parsers map[string]*regexp.Regexp = map[string]*regexp.Regexp{}

//var timeExtractors map[string]func(map[string] string) time.Time = map[string]func(map[string] string) time.Time{"landsat":landsatTime, "modis1": modisTime, "modis2": modisTime}

func init() {
	for key, value := range parserStrings {
		parsers[key] = regexp.MustCompile(value)
	}
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
		t := time.Date(year, 0, 0, 0, 0, 0, 0, time.UTC)
		if _, ok := nameFields["julian_day"]; ok {
			julianDay, _ := strconv.Atoi(nameFields["julian_day"])
			t = t.Add(time.Hour * 24 * time.Duration(julianDay))
		}
		if _, ok := nameFields["month"]; ok {
			if _, ok := nameFields["day"]; ok {
				month, _ := strconv.Atoi(nameFields["month"])
				day, _ := strconv.Atoi(nameFields["day"])
				t_help := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
				t = t.Add(time.Hour * 24 * time.Duration(t_help.YearDay()))
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

func main() {

	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	c := session.DB("test").C("himawari8")
	//c := session.DB("test").C("mcd1")
	//c := session.DB("test").C("mcd43a2")
	//c := session.DB("test").C("mcd43a4")
	//c := session.DB("test").C("agdcv2")

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		parts := strings.Split(s.Text(), "\t")
		if len(parts) != 2 {
			fmt.Printf("Input not recognised: %s\n", s.Text())
		}

		gdalFile := geolib.GDALFile{}
		err := json.Unmarshal([]byte(parts[1]), &gdalFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		geoFile := GeoFile{FileName: parts[0], Driver: gdalFile.Driver}

		nameFields, timeStamp := parseName(parts[0])

		if nameFields != nil {
			
			for _, ds := range gdalFile.DataSets {
				if ds.ProjWKT != "" {
					poly := geolib.GetPolygon(ds.ProjWKT, ds.GeoTransform, ds.XSize, ds.YSize)
					polyWGS84 := poly.ReprojectToWGS84()

					var times []time.Time
					if nc_times, ok := ds.Extras["nc_times"]; ok {
						for _, timestr := range nc_times {
							t, err := time.Parse("2006-01-02T15:04:05Z", timestr)
							if err != nil {
								fmt.Println(err)
							}
							times = append(times, t)
						}
					} else {
						times = []time.Time{timeStamp}
					}

					geoFile.DataSets = append(geoFile.DataSets, GeoMetaData{DataSetName: ds.DataSetName, TimeStamps: times, FileNameFields: nameFields, Location: polyWGS84.AsPolygon(), RasterCount: ds.RasterCount, Type: ds.Type, XSize: ds.XSize, YSize: ds.YSize, ProjWKT: ds.ProjWKT, GeoTransform: ds.GeoTransform})
				}
			}

			/*
			err = c.Insert(&geoFile)
			if err != nil {
				log.Fatal(err)
			}
			*/
				out, err := json.Marshal(&geoFile)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Printf("%s\n", string(out))
		} else {
			log.Printf("%s non parseable", parts[0])
		}
	}
}
