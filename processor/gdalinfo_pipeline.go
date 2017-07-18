package processor

import (
	"fmt"
	"context"
	"regexp"
)

type InfoPipeline struct {
	Context context.Context
	Error   chan error
	RPCAddress string
}

func InitInfoPipeline(ctx context.Context, rpcAddr string, errChan chan error) (*InfoPipeline) {
	return &InfoPipeline{
		Context:   ctx,
		Error:   errChan,
		RPCAddress: rpcAddr,
	}
}

func (dp *InfoPipeline) Process(rootPath string, contains *regexp.Regexp) chan struct{} {
	i := NewFileCrawler(rootPath, contains, dp.Error)
	go func() {
		i.In <- rootPath
		close(i.In)
	}()

	p := NewJSONPrinter(dp.Error)

	grpcInfo := NewInfoGRPC(dp.Context, dp.RPCAddress, dp.Error)
	if grpcInfo == nil {
		dp.Error <- fmt.Errorf("Couldn't instantiate RPCTiler %s/n", dp.RPCAddress)
		close(p.Out)
		return p.Out
	}

	grpcInfo.In = i.Out
	p.In = grpcInfo.Out

	go i.Run()
	go grpcInfo.Run()
	go p.Run()

	return p.Out
}
