package main


import (
	proc "./processor"
	"fmt"
	"context"
)

func main() {
	ctx := context.Background()
	ctx, ctxCancel := context.WithCancel(ctx)
	errChan := make(chan error)
	tp := proc.InitInfoPipeline(ctx, "localhost:6060", errChan)
	select {
	case <- tp.Process("./", nil):
		fmt.Println("Pipeline Total Time")
	case err := <- errChan:
		fmt.Println(err)
		ctxCancel()
	case <- ctx.Done():
		fmt.Println("GGGG")
	}
}