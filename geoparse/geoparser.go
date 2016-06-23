package main

import (
	"../geolib"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GeoMetaData struct {
	DataSetName  string    `json:"ds_name"`
	TimeStamps      []string         `json:"timestamps"`
	FileNameFields map[string]string `json:"filename_fields"`
	Polygon        json.RawMessage   `json:"polygon"`
	RasterCount  int       `json:"raster_count"`
	Type           string            `json:"array_type"`
	XSize          int               `json:"x_size"`
	YSize          int               `json:"y_size"`
	ProjWKT        string            `json:"proj_wkt"`
	GeoTransform   []float64         `json:"geotransform"`
}

type GeoFile struct {
	Driver   string        `json:"file_type"`
	DataSets []GeoMetaData `json:"geo_metadata"`
}

var parserStrings map[string]string = map[string]string{"landsat": `LC(?P<mission>\d)(?P<path>\d\d\d)(?P<row>\d\d\d)(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<processing_level>[a-zA-Z0-9]+)_(?P<band>[a-zA-Z0-9]+)`,
				  "modis1": `M(?P<satellite>[OD|YD])(?P<product>[0-9]+_[A-Z0-9]+).A[0-9]+.[0-9]+.(?P<collection_version>\d\d\d).(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)`,
				  "modis2": `MCD43A4.A[0-9]+.(?P<horizontal>h\d\d)(?P<vertical>v\d\d).(?P<resolution>\d\d\d).(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)`,
				  "agdc_landsat1": `LS(?P<mission>\d)_(?P<sensor>[A-Z]+)_(?P<correction>[A-Z]+)_(?P<epsg>\d+)_(?P<x_coord>-?\d+)_(?P<y_coord>-?\d+)_(?P<year>\d\d\d\d).`,}

var parsers map[string]*regexp.Regexp = map[string]*regexp.Regexp{}
//var timeExtractors map[string]func(map[string] string) time.Time = map[string]func(map[string] string) time.Time{"landsat":landsatTime, "modis1": modisTime, "modis2": modisTime}

func init() {
	for key, value := range(parserStrings) {
		parsers[key] = regexp.MustCompile(value)
	}
}

func parseName(filePath string) (map[string]string, time.Time) {

	for _, r := range(parsers) {
		_, fileName := filepath.Split(filePath)

		if (r.MatchString(fileName)) {
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
		
		geoFile := GeoFile{Driver: gdalFile.Driver}

		nameFields, timeStamp := parseName(parts[0])

		for _, ds := range gdalFile.DataSets {
			if ds.ProjWKT != "" {
				poly := geolib.GetPolygon(ds.ProjWKT, ds.GeoTransform, ds.XSize, ds.YSize)
				polyWGS84 := poly.ReprojectToWGS84()

				var times []string
				if nc_times, ok := ds.Extras["nc_times"]; ok {
					times = nc_times
				} else {
					times = []string{timeStamp.Format("2006-01-02T15:04:05Z")}
				}

				geoFile.DataSets = append(geoFile.DataSets, GeoMetaData{DataSetName: ds.DataSetName, TimeStamps: times, FileNameFields: nameFields, Polygon: json.RawMessage(polyWGS84.ToGeoJSON()), RasterCount: ds.RasterCount, Type: ds.Type, XSize: ds.XSize, YSize: ds.YSize, ProjWKT: ds.ProjWKT, GeoTransform: ds.GeoTransform})

			}
		}
		out, err := json.Marshal(&geoFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%s\tgdal\t%s\n", parts[0], string(out))
		
	}
}
