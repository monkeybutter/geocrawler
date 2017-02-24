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
	NextContinuationToken *string
	KeyCount              int
	MaxKeys               int
	IsTruncated           bool
	Contents              []Contents
}

func ListBuckets(contToken, startAfter *string) (ListBucketResult, error) {
	baseURL := "http://landsat-pds.s3.amazonaws.com?list-type=2"

	if startAfter != nil {
		baseURL = fmt.Sprintf("%s&start-after=%s", baseURL, url.QueryEscape(*startAfter))
	}
	if contToken != nil {
		baseURL = fmt.Sprintf("%s&continuation-token=%s", baseURL, url.QueryEscape(*contToken))
	}
	resp, err := http.Get(baseURL)
	if err != nil {
		return ListBucketResult{IsTruncated: false}, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	var lbRes ListBucketResult
	err = xml.Unmarshal(buf.Bytes(), &lbRes)
	if err != nil {
		return ListBucketResult{IsTruncated: false}, err
	}

	return lbRes, nil

}

type S3Crawler struct {
	Out   chan string
	Error chan error
	startAfter *string
}

func NewS3Crawler(stAfter *string, errChan chan error) *S3Crawler {
	return &S3Crawler{
		Out: make(chan string, 100),
		Error: errChan,
		startAfter: stAfter,
	}
}

func (fc *S3Crawler) Run() {
	defer close(fc.Out)

	lbRes := ListBucketResult{IsTruncated: true, NextContinuationToken:nil}
	var err error

	for lbRes.IsTruncated {
		lbRes, err = ListBuckets(lbRes.NextContinuationToken, fc.startAfter)
		if err != nil {
			fc.Error <- err
			continue
		}
		for _, contents := range lbRes.Contents {
			if filepath.Ext(contents.Key) == ".TIF" {
				fc.Out <- filepath.Join("/vsis3/landsat-pds/", contents.Key)
			}
		}
		fc.startAfter = nil
	}
}
