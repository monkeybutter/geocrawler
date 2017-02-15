package processor

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
)

type Contents struct {
	Key          string
	LastModified string
	ETag         string
	Size         int
	StorageClass string
}

type ListBucketResult struct {
	Name                  string
	NextContinuationToken string
	KeyCount              int
	MaxKeys               int
	IsTruncated           bool
	Contents              []Contents
}

func ListBuckets(contToken *string) (ListBucketResult, error) {
	baseURL := "http://landsat-pds.s3.amazonaws.com?list-type=2"

	if contToken != nil {
		baseURL = fmt.Sprintf("%s&continuation-token=%s", baseURL, url.QueryEscape(*contToken))
	}
	fmt.Println(baseURL)
	resp, err := http.Get(baseURL)
	if err != nil {
		return ListBucketResult{}, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	var lbRes ListBucketResult
	err = xml.Unmarshal(buf.Bytes(), &lbRes)
	if err != nil {
		return ListBucketResult{}, err
	}

	return lbRes, nil

}

func main() {
	fmt.Println("AWS Buckets")

	lbRes, err := ListBuckets(nil)
	if err != nil {
		return
	}
	for _, contents := range lbRes.Contents {
		ext := filepath.Ext(contents.Key)
		if ext == ".TIF" || ext == ".ovr" {
			fmt.Println(filepath.Join("/vsis3/landsat-pds/", contents.Key))
		}
	}
	lbRes, err = ListBuckets(&lbRes.NextContinuationToken)
	for _, contents := range lbRes.Contents {
		ext := filepath.Ext(contents.Key)
		if ext == ".TIF" || ext == ".ovr" {
			fmt.Println(filepath.Join("/vsis3/landsat-pds/", contents.Key))
		}
	}
}
