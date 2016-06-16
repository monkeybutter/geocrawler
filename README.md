# geocrawler
A crawler for geospatial files

###About:
**geocrawler** is a command line tool to extract geospatial metadata in JSON format.

A file or directory is passed to **geocrawler** as an argument returning a JSON document for the file or a list of JSON documents for the geospatial files under that directory (the crawler traverses the whole tree under the specified directory)

There is a flag _-c_ to determine the level of concurrency of the crawler. This is intended to speed up traversing directories with a large number of files by using several cores in the machine. _-c_ should be set to a number equal or lower to the number of cores in the machine doing the crawling.

This program is written in Go wrapping the C GDAL library. There is a linux precompiled version available at the bin directory.

###Examples:

Extracting metadata for one file:
```bash
geocrawler LC80640052015252LGN00_B1.TIF
```
```
{"file_type":"GTiff","datasets":[{"ds_name":"/g/data1/sp9/earthengine-public/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B1.TIF","raster_count":1,"array_type":"Uint16","x_size":9141,"y_size":9161,"proj4":"+proj=utm +zone=11 +datum=WGS84 +units=m +no_defs ","geotransform":[344085,30,0,8.689215e+06,0,-30]}]}
```

Extracting metadata for all the files under a directory:
```
geocrawler LC80640052015252LGN00/
```
```
/g/data1/sp9/earthengine-public/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B1.TIF	{"file_type":"GTiff","datasets":[{"ds_name":"/g/data1/sp9/earthengine-public/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B1.TIF","raster_count":1,"array_type":"Uint16","x_size":9141,"y_size":9161,"proj4":"+proj=utm +zone=11 +datum=WGS84 +units=m +no_defs ","geotransform":[344085,30,0,8.689215e+06,0,-30]}]}
/g/data1/sp9/earthengine-public/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B2.TIF	{"file_type":"GTiff","datasets":[{"ds_name":"/g/data1/sp9/earthengine-public/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B2.TIF","raster_count":1,"array_type":"Uint16","x_size":9141,"y_size":9161,"proj4":"+proj=utm +zone=11 +datum=WGS84 +units=m +no_defs ","geotransform":[344085,30,0,8.689215e+06,0,-30]}]}
/g/data1/sp9/earthengine-public/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B3.TIF	{"file_type":"GTiff","datasets":[{"ds_name":"/g/data1/sp9/earthengine-public/landsat/L8/064/005/LC80640052015252LGN00/LC80640052015252LGN00_B3.TIF","raster_count":1,"array_type":"Uint16","x_size":9141,"y_size":9161,"proj4":"+proj=utm +zone=11 +datum=WGS84 +units=m +no_defs ","geotransform":[344085,30,0,8.689215e+06,0,-30]}]}
...
```

Same example but using 4 processes extracting the metadata concurrently:
```
geocrawler -c 4 LC80640052015252LGN00/
```


*This program has been developed at the NCI*
