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
	TimeStamp      time.Time         `json:"timestamp"`
	FileNameFields map[string]string `json:"filename_fields"`
	Polygon        json.RawMessage   `json:"polygon"`
	Type           string            `json:"array_type"`
	XSize          int               `json:"x_size"`
	YSize          int               `json:"y_size"`
	ProjWKT        string            `json:"proj_wkt"`
	GeoTransform   []float64         `json:"geotransform"`
}

var parserStrings map[string]string = map[string]string{"landsat": `LC(?P<mission>\d)(?P<path>\d\d\d)(?P<row>\d\d\d)(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<processing_level>[a-zA-Z0-9]+)_(?P<band>[a-zA-Z0-9]+)`,
				  "modis1": `M(?P<satellite>[OD|YD])(?P<product>[0-9]+_[A-Z0-9]+).A[0-9]+.[0-9]+.(?P<collection_version>\d\d\d).(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)`,
				  "modis2": `MCD43A4.A[0-9]+.(?P<horizontal>h\d\d)(?P<vertical>v\d\d).(?P<resolution>\d\d\d).(?P<year>\d\d\d\d)(?P<julian_day>\d\d\d)(?P<hour>\d\d)(?P<minute>\d\d)(?P<second>\d\d)`,}

var parsers map[string]*regexp.Regexp = map[string]*regexp.Regexp{}
var timeExtractors map[string]func(map[string] string) time.Time = map[string]func(map[string] string) time.Time{"landsat":landsatTime, "modis1": modisTime, "modis2": modisTime}

func start() {
	for key, value := range(parserStrings) {
		parsers[key] = regexp.MustCompile(value)
	}
}

func parseName(filePath string) (map[string]string, time.Time) {

	for parserName, r := range(parsers) {
		_, fileName := filepath.Split(filePath)

		if (r.MatchString(fileName)) {
			match := r.FindStringSubmatch(fileName)

			result := make(map[string]string)
			for i, name := range r.SubexpNames() {
				if i != 0 {
					result[name] = match[i]
				}
			}

			return result, timeExtractors[parserName](result)
		}	
	}
	return nil, time.Time{}
}

func landsatTime(nameFields map[string]string) time.Time {
	year, _ := strconv.Atoi(nameFields["year"])
	julianDay, _ := strconv.Atoi(nameFields["julian_day"])
	t := time.Date(year, 0, 0, 0, 0, 0, 0, time.UTC)
	t = t.Add(time.Hour * 24 * time.Duration(julianDay))

	return t
}

func modisTime(nameFields map[string]string) time.Time {
	year, _ := strconv.Atoi(nameFields["year"])
	julianDay, _ := strconv.Atoi(nameFields["julian_day"])
	hour, _ := strconv.Atoi(nameFields["hour"])
	minute, _ := strconv.Atoi(nameFields["minute"])
	second, _ := strconv.Atoi(nameFields["second"])

	t := time.Date(year, 0, 0, 0, 0, 0, 0, time.UTC)
	t = t.Add(time.Hour * 24 * time.Duration(julianDay))
	t = t.Add(time.Hour * time.Duration(hour))
	t = t.Add(time.Minute * time.Duration(minute))
	t = t.Add(time.Second * time.Duration(second))

	return t

}

func main() {
	start()
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

		nameFields, timeStamp := parseName(parts[0])

		for _, ds := range gdalFile.DataSets {
			poly := geolib.GetPolygon(ds.ProjWKT, ds.GeoTransform, ds.XSize, ds.YSize)
			polyWGS84 := poly.ReprojectToWGS84()

			fileMetaData := GeoMetaData{TimeStamp: timeStamp, FileNameFields: nameFields, Polygon: json.RawMessage(polyWGS84.ToGeoJSON()),
				Type: ds.Type, XSize: ds.XSize, YSize: ds.YSize, ProjWKT: ds.ProjWKT, GeoTransform: ds.GeoTransform}

			out, err := json.Marshal(&fileMetaData)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("%s\t%s\n", parts[0], string(out))

		}
	}
}
