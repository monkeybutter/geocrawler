package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "./gdalservice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	_ "net/http/pprof"
	"net/http"
)

type server struct {
	Pool *pb.ProcessPool
}

func (s *server) Info(ctx context.Context, in *pb.GeoRequest) (*pb.GeoFile, error) {
	rChan := make(chan *pb.GeoFile)
	defer close(rChan)
	errChan := make(chan error)
	defer close(errChan)

	s.Pool.AddInfoQueue(&pb.InfoTask{in, rChan, errChan})

	select {
	case out, ok := <-rChan:
		if !ok {
			return &pb.GeoFile{}, fmt.Errorf("task response channel has been closed")
		}
		return out, nil
	case err := <-errChan:
		return &pb.GeoFile{}, fmt.Errorf("Error in ops: %v", err)
	}
}

func main() {
	port := flag.Int("p", 6000, "gRPC server listening port.")
	poolSize := flag.Int("n", 4, "Maximum number of requests handled concurrently.")
	//aws := flag.Bool("aws", true, "Needs access to AWS S3 rasters?")
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	p := pb.CreateProcessPool(*poolSize)

	s := grpc.NewServer()
	pb.RegisterGDALServer(s, &server{Pool: p})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
