package main

import (
	proc "./processor"
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	ctx, ctxCancel := context.WithCancel(ctx)
	errChan := make(chan error)
	tp := proc.InitInfoPipeline(ctx, "localhost:6060", errChan)
	select {
	case <-tp.Process("/g/data2/u39/public/data/modis/lpdaac-tiles-c5/MCD12Q1.051/2013.01.01/", ".*"):
		fmt.Println("Pipeline Total Time")
	case err := <-errChan:
		fmt.Println(err)
		ctxCancel()
	case <-ctx.Done():
		fmt.Println("GGGG")
	}
}
