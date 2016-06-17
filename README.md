# geocrawler
A crawler for geospatial files

###About:
**geocrawler** is a command line tool to extract geospatial metadata in JSON format.
**geoparser** is a command line tool to process the output of **geocrawler** and generate an indexable JSON representation of a file containing geospatial data. **geoparser** read from standard input and is intended to be used as a piped command after **geocrawler**.

A file or directory is passed to **geocrawler** as an argument returning a JSON document for the file or a list of JSON documents for the geospatial files under that directory (the crawler traverses the whole tree under the specified directory)

There is a flag _-c_ to determine the level of concurrency of the crawler. This is intended to speed up traversing directories with a large number of files by using several cores in the machine. _-c_ should be set to a number equal or lower to the number of cores in the machine doing the crawling.

This program is written in Go wrapping the C GDAL library. There is a linux precompiled version available at the bin directory.

###Examples:

Extracting metadata for one file:
```bash
geocrawler LC80640052015252LGN00_B1.TIF
```
```
{"file_type":"GTiff","datasets":[{"ds_name":"/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B1.TIF","raster_count":1,"array_type":"Uint16","x_size":9141,"y_size":9161,"proj4":"+proj=utm +zone=11 +datum=WGS84 +units=m +no_defs ","geotransform":[344085,30,0,8.689215e+06,0,-30]}]}
```

Extracting metadata for all the files under a directory:
```
geocrawler LC80640052015252LGN00/
```
```
/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B1.TIF	{"file_type":"GTiff","datasets":[{"ds_name":"/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B1.TIF","raster_count":1,"array_type":"Uint16","x_size":9141,"y_size":9161,"proj4":"+proj=utm +zone=11 +datum=WGS84 +units=m +no_defs ","geotransform":[344085,30,0,8.689215e+06,0,-30]}]}
/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B2.TIF	{"file_type":"GTiff","datasets":[{"ds_name":"/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B2.TIF","raster_count":1,"array_type":"Uint16","x_size":9141,"y_size":9161,"proj4":"+proj=utm +zone=11 +datum=WGS84 +units=m +no_defs ","geotransform":[344085,30,0,8.689215e+06,0,-30]}]}
/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B3.TIF	{"file_type":"GTiff","datasets":[{"ds_name":"/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B3.TIF","raster_count":1,"array_type":"Uint16","x_size":9141,"y_size":9161,"proj4":"+proj=utm +zone=11 +datum=WGS84 +units=m +no_defs ","geotransform":[344085,30,0,8.689215e+06,0,-30]}]}
...
```

Same example but using 4 processes extracting the metadata concurrently and writing the output to a file (to be later ingested in a database):
```
geocrawler -c 4 LC80640052015252LGN00/ > out.tsv
```

Extracting metadata for all the files under a directory and parsing into an indexable object:
```
geocrawler LC80640052015252LGN00/ | geoparser
```
```
/landsat/L8/102/078/LC81020782015199LGN00/LC81020782015199LGN00_B1.TIF	{"timestamp":"2015-06-17T00:00:00Z","filename_fields":{"band":"B1","julian_day":"199","mission":"8","path":"102","processing_level":"LGN00","row":"078","year":"2015"},"polygon":{"type":"Polygon","coordinates":[[[132.428274005358645,-24.930273120739798],[132.382199681744368,-27.027829648222031],[134.692598341469136,-27.051847101496204],[134.698015969944748,-24.952165411227803],[132.428274005358645,-24.930273120739798]]]},"array_type":"Uint16","x_size":7641,"y_size":7751,"proj_wkt":"PROJCS[\"WGS 84 / UTM zone 53N\",GEOGCS[\"WGS 84\",DATUM[\"WGS_1984\",SPHEROID[\"WGS 84\",6378137,298.257223563,AUTHORITY[\"EPSG\",\"7030\"]],AUTHORITY[\"EPSG\",\"6326\"]],PRIMEM[\"Greenwich\",0,AUTHORITY[\"EPSG\",\"8901\"]],UNIT[\"degree\",0.0174532925199433,AUTHORITY[\"EPSG\",\"9122\"]],AUTHORITY[\"EPSG\",\"4326\"]],PROJECTION[\"Transverse_Mercator\"],PARAMETER[\"latitude_of_origin\",0],PARAMETER[\"central_meridian\",135],PARAMETER[\"scale_factor\",0.9996],PARAMETER[\"false_easting\",500000],PARAMETER[\"false_northing\",0],UNIT[\"metre\",1,AUTHORITY[\"EPSG\",\"9001\"]],AXIS[\"Easting\",EAST],AXIS[\"Northing\",NORTH],AUTHORITY[\"EPSG\",\"32653\"]]","geotransform":[240285,30,0,-2.759685e+06,0,-30]}
/landsat/L8/102/078/LC81020782015199LGN00/LC81020782015199LGN00_B2.TIF	{"timestamp":"2015-06-17T00:00:00Z","filename_fields":{"band":"B2","julian_day":"199","mission":"8","path":"102","processing_level":"LGN00","row":"078","year":"2015"},"polygon":{"type":"Polygon","coordinates":[[[132.428274005358645,-24.930273120739798],[132.382199681744368,-27.027829648222031],[134.692598341469136,-27.051847101496204],[134.698015969944748,-24.952165411227803],[132.428274005358645,-24.930273120739798]]]},"array_type":"Uint16","x_size":7641,"y_size":7751,"proj_wkt":"PROJCS[\"WGS 84 / UTM zone 53N\",GEOGCS[\"WGS 84\",DATUM[\"WGS_1984\",SPHEROID[\"WGS 84\",6378137,298.257223563,AUTHORITY[\"EPSG\",\"7030\"]],AUTHORITY[\"EPSG\",\"6326\"]],PRIMEM[\"Greenwich\",0,AUTHORITY[\"EPSG\",\"8901\"]],UNIT[\"degree\",0.0174532925199433,AUTHORITY[\"EPSG\",\"9122\"]],AUTHORITY[\"EPSG\",\"4326\"]],PROJECTION[\"Transverse_Mercator\"],PARAMETER[\"latitude_of_origin\",0],PARAMETER[\"central_meridian\",135],PARAMETER[\"scale_factor\",0.9996],PARAMETER[\"false_easting\",500000],PARAMETER[\"false_northing\",0],UNIT[\"metre\",1,AUTHORITY[\"EPSG\",\"9001\"]],AXIS[\"Easting\",EAST],AXIS[\"Northing\",NORTH],AUTHORITY[\"EPSG\",\"32653\"]]","geotransform":[240285,30,0,-2.759685e+06,0,-30]}
/landsat/L8/102/078/LC81020782015199LGN00/LC81020782015199LGN00_B3.TIF	{"timestamp":"2015-06-17T00:00:00Z","filename_fields":{"band":"B3","julian_day":"199","mission":"8","path":"102","processing_level":"LGN00","row":"078","year":"2015"},"polygon":{"type":"Polygon","coordinates":[[[132.428274005358645,-24.930273120739798],[132.382199681744368,-27.027829648222031],[134.692598341469136,-27.051847101496204],[134.698015969944748,-24.952165411227803],[132.428274005358645,-24.930273120739798]]]},"array_type":"Uint16","x_size":7641,"y_size":7751,"proj_wkt":"PROJCS[\"WGS 84 / UTM zone 53N\",GEOGCS[\"WGS 84\",DATUM[\"WGS_1984\",SPHEROID[\"WGS 84\",6378137,298.257223563,AUTHORITY[\"EPSG\",\"7030\"]],AUTHORITY[\"EPSG\",\"6326\"]],PRIMEM[\"Greenwich\",0,AUTHORITY[\"EPSG\",\"8901\"]],UNIT[\"degree\",0.0174532925199433,AUTHORITY[\"EPSG\",\"9122\"]],AUTHORITY[\"EPSG\",\"4326\"]],PROJECTION[\"Transverse_Mercator\"],PARAMETER[\"latitude_of_origin\",0],PARAMETER[\"central_meridian\",135],PARAMETER[\"scale_factor\",0.9996],PARAMETER[\"false_easting\",500000],PARAMETER[\"false_northing\",0],UNIT[\"metre\",1,AUTHORITY[\"EPSG\",\"9001\"]],AXIS[\"Easting\",EAST],AXIS[\"Northing\",NORTH],AUTHORITY[\"EPSG\",\"32653\"]]","geotransform":[240285,30,0,-2.759685e+06,0,-30]}
```

*This program has been developed at the NCI*
