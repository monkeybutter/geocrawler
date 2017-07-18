package processor

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "../grpc/gdalservice"
)

type GeoInfoGRPC struct {
	Context context.Context
	In     chan *pb.GeoRequest
	Out    chan *pb.GeoFile
	Error  chan error
	Client string
}

func NewInfoGRPC(ctx context.Context, serverAddress string, errChan chan error) *GeoInfoGRPC {
	return &GeoInfoGRPC{
		Context: ctx,
		In:     make(chan *pb.GeoRequest, 100),
		Out:    make(chan *pb.GeoFile, 100),
		Error:  errChan,
		Client: serverAddress,
	}
}

func (gi *GeoInfoGRPC) Run() {
	defer close(gi.Out)

	conn, err := grpc.Dial(gi.Client, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("gRPC connection problem: %v", err)
	}
	defer conn.Close()
	for gran := range gi.In {
		select {
		case <-gi.Context.Done():
			gi.Error <- fmt.Errorf("Tile gRPC context has been cancel: %v", gi.Context.Err())
			return
		default:
			go func(g *pb.GeoRequest) {
				c := pb.NewGDALClient(conn)
				r, err := c.Info(gi.Context, g)
				if err != nil {
					gi.Error <- err
					r = &pb.GeoFile{}
				}
				gi.Out <- r
			}(gran)
		}
	}
}

