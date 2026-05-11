package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/university/sports-event-planner-platform/pkg/health"
)

type Clients struct {
	conns map[string]*grpc.ClientConn
}

func Dial(_ context.Context, endpoints map[string]string) (*Clients, error) {
	conns := make(map[string]*grpc.ClientConn, len(endpoints))
	for name, addr := range endpoints {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		conns[name] = conn
	}
	return &Clients{conns: conns}, nil
}

func (c *Clients) Checks() map[string]health.Checker {
	checks := make(map[string]health.Checker, len(c.conns))
	for name, conn := range c.conns {
		conn := conn
		checks["grpc_"+name] = func(context.Context) error {
			if conn.GetState() == connectivity.Shutdown {
				return errors.New("grpc connection is shutdown")
			}
			return nil
		}
	}
	return checks
}

func (c *Clients) Close() error {
	var err error
	for _, conn := range c.conns {
		err = errors.Join(err, conn.Close())
	}
	return err
}
