package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"

	pb "./gdalinfo"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

func (s *server) Extract(ctx context.Context, in *pb.Request) (*pb.GDALFile, error) {
	res := &pb.GDALFile{}

	var outb, errb bytes.Buffer
	cmd := exec.Command("./c_warp/warp", pb.GDALFile.FilePath)
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		log.Println("Error running command:", err)
		return res, err
	}

	err = proto.Unmarshal(outb.Bytes(), res)
	if err != nil {
		log.Println("Error deserialising proto message:", err)
		return res, err
	}

	if len(errb.Bytes()) > 0 {
		err = fmt.Errorf(string(errb.Bytes()))
		log.Println("Error returned by process", err)
		return res, err
	}

	return res, err
}

func main() {
	port := flag.Int("p", 6000, "gRPC server listening port.")
	conc := flag.Int("c", 24, "Maximum number of connections handled concurrently.")
	//aws := flag.Bool("aws", true, "Needs access to AWS S3 rasters?")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	lis = netutil.LimitListener(lis, *conc)

	s := grpc.NewServer()
	pb.RegisterGDALInfoServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
