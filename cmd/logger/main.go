package main

import (
	"context"
	"flag"
	pbCDC "github.com/amasynikov/grpc-webinar/internal/genproto/cdc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

var (
	storage  = flag.String("storage", "0.0.0.0:8081", "CDC service address")
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc, err := grpc.DialContext(ctx, *storage,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (_ net.Conn, err error) {
			log.Trace().Str("address", s).Msg("dialing to storage-service")
			defer func() {
				if err != nil {
					log.Error().Caller().Str("address", s).Err(err).Msg("dial to storage-service failed")
				} else {
					log.Info().Caller().Str("address", s).Msg("dial to storage-service done")
				}
			}()
			return net.Dial("tcp", s)
		}),
	)
	if err != nil {
		log.Fatal().Caller().Str("storage", *storage).Err(err).Msg("dial failed")
		return
	}

	client := pbCDC.NewCDCClient(cc)

	stream, err := client.Listen(ctx, &pbCDC.ListenRequest{})
	if err != nil {
		log.Fatal().Caller().Str("storage", *storage).Err(err).Msg("dial failed")
		return
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatal().Caller().Str("storage", *storage).Err(err).Msg("dial failed")
			return
		}
		log.Info().Caller().Str("event", msg.GetEvent().String()).Str("id", msg.GetData().GetId()).Bytes("data", msg.GetData().GetRaw()).Msg("")
	}
}
