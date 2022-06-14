package main

import (
	"context"
	"flag"
	"github.com/amasynikov/grpc-webinar/internal/storage"
	"net"
	"net/url"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	pbCDC "github.com/amasynikov/grpc-webinar/internal/genproto/cdc"
	pbCRUD "github.com/amasynikov/grpc-webinar/internal/genproto/crud"
)

var (
	socket   = flag.String("socket", "tcp://0.0.0.0:8081", "socket of auth service")
	logLevel = flag.String("log-level", "info", "logging level")
)

func init() {
	zerolog.TimeFieldFormat = "2006.01.02-15:04:05.000"
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	flag.Parse()

	l, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return
	}
	zerolog.SetGlobalLevel(l)

	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
					now := time.Now()
					res, err := handler(ctx, req)
					if err != nil {
						log.Error().Caller().Err(err).Str("method", info.FullMethod).Stringer("duration", time.Since(now)).Msg("")
					} else {
						log.Trace().Caller().Str("method", info.FullMethod).Stringer("duration", time.Since(now)).Msg("")
					}
					return res, err
				},
			),
		),
	)

	storage := storage.New()

	pbCRUD.RegisterCRUDServer(s, storage)
	pbCDC.RegisterCDCServer(s, storage)

	url, err := url.Parse(*socket)
	if err != nil {
		log.Fatal().Caller().Str("socket", *socket).Err(err).Msg("")
	}

	listen, err := net.Listen(url.Scheme, url.Host)
	if err != nil {
		log.Fatal().Caller().Str("socket", *socket).Err(err).Msg("")
		return
	}

	if err := s.Serve(listen); err != nil {
		log.Fatal().Caller().Str("socket", *socket).Err(err).Msg("")
		return
	}
}
