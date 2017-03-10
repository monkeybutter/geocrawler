#include <stdio.h>
#include <stdlib.h>
#include "gdal.h"
#include "ogr_api.h"
#include "ogr_srs_api.h"
#include "cpl_conv.h"

int getParamsFromGeoTransform(char *projWKT, double geoTrans[6], int xSize, int ySize, char *result[]);

int main(int argc, char **argv)
{
	char *projWKT = "PROJCS[\"WGS 84 / UTM zone 29N\",GEOGCS[\"WGS 84\",DATUM[\"WGS_1984\",SPHEROID[\"WGS 84\",6378137,298.257223563,AUTHORITY[\"EPSG\",\"7030\"]],AUTHORITY[\"EPSG\",\"6326\"]],PRIMEM[\"Greenwich\",0,AUTHORITY[\"EPSG\",\"8901\"]],UNIT[\"degree\",0.0174532925199433,AUTHORITY[\"EPSG\",\"9122\"]],AUTHORITY[\"EPSG\",\"4326\"]],PROJECTION[\"Transverse_Mercator\"],PARAMETER[\"latitude_of_origin\",0],PARAMETER[\"central_meridian\",-9],PARAMETER[\"scale_factor\",0.9996],PARAMETER[\"false_easting\",500000],PARAMETER[\"false_northing\",0],UNIT[\"metre\",1,AUTHORITY[\"EPSG\",\"9001\"]],AXIS[\"Easting\",EAST],AXIS[\"Northing\",NORTH],AUTHORITY[\"EPSG\",\"32629\"]]";
	double geot[6] = {331485, 30, 0, 8.920215e+06, 0, -30};
	int xSize = 9131;
	int ySize = 9121;
	char *result[3];

	while(1) 
	{
		getParamsFromGeoTransform(projWKT, geot, xSize, ySize, result);
		printf("%s %s %s\n", result[0], result[1], result[2]);
	}

	return 0;
}

int getParamsFromGeoTransform(char *projWKT, double geoTrans[6], int xSize, int ySize, char *result[])
{
	double ulX, ulY, lrX, lrY;
	char *polyWKT = malloc(500);
	char *pWKT;
	char *ppszSrcText;
	char *pszProj4;
	char *pszProjWKT;
	OGRSpatialReferenceH hSRS;
	OGRGeometryH hPt;

	GDALApplyGeoTransform(geoTrans, 0, 0, &ulX, &ulY);
	GDALApplyGeoTransform(geoTrans, xSize, ySize, &lrX, &lrY);

	sprintf(polyWKT, "POLYGON ((%f %f,%f %f,%f %f,%f %f,%f %f))", ulX, ulY, ulX, lrY, lrX, lrY, lrX, ulY, ulX, ulY);

	hSRS = OSRNewSpatialReference(projWKT);

	pWKT = polyWKT;
	OGR_G_CreateFromWkt(&pWKT, hSRS, &hPt);
	
	OGR_G_ExportToWkt(hPt, &ppszSrcText);

	OSRExportToProj4(hSRS, &pszProj4);
	
	OSRExportToWkt(hSRS, &pszProjWKT);
	
	result[0] = ppszSrcText;
	result[1] = pszProj4;
	result[2] = pszProjWKT;

	OGR_G_DestroyGeometry(hPt);
	OSRDestroySpatialReference(hSRS);
	free(polyWKT);
	CPLFree(ppszSrcText);
	CPLFree(pszProj4);
	CPLFree(pszProjWKT);
	
	return 0; 
}
