package grpc_auto

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"log"
	"time"
)

type AutoGrpc struct {
	conn *grpc.ClientConn
	target string
	opts []grpc.DialOption
	config *Config
}

type Config struct {
	RetryInterval time.Duration
	Asynchronous  bool
}

func waitConnection(target string, wait time.Duration, opts ... grpc.DialOption) *grpc.ClientConn {
	for {
		conn, err := grpc.Dial(target, opts ...)
		if err == nil {
			return conn
		}
		log.Println(err)
		time.Sleep(wait)
	}
}

func getConn (auto *AutoGrpc) *grpc.ClientConn {
	if auto.conn == nil || auto.conn.GetState() != connectivity.Ready {
		if !auto.config.Asynchronous {
			auto.conn = waitConnection(auto.target, auto.config.RetryInterval, auto.opts ...)
		}
	}
	return auto.conn
}

func (a *AutoGrpc) GetConnection() *grpc.ClientConn {
	return getConn(a)
}

func NewAutoGrpc(target string, config *Config, opts ... grpc.DialOption) *AutoGrpc {
	return &AutoGrpc{target:target, config: config, opts: opts}
}