package main

import (
	"context"
	"flag"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbAuth "github.com/amasynikov/grpc-webinar/internal/genproto/auth"
	pbCRUD "github.com/amasynikov/grpc-webinar/internal/genproto/crud"
	"github.com/amasynikov/grpc-webinar/internal/web"
)

var (
	storage  = flag.String("storage", "0.0.0.0:8081", "CRUD-service address")
	auth     = flag.String("auth", "0.0.0.0:8080", "Auth-service address")
	logLevel = flag.String("log-level", "info", "Logging level")
	port     = flag.Int("port", 80, "Web-service port")
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

	ccAuth, err := grpc.DialContext(ctx, *auth,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (_ net.Conn, err error) {
			log.Trace().Str("address", s).Msg("dialing to auth-service")
			defer func() {
				if err != nil {
					log.Error().Caller().Str("address", s).Err(err).Msg("dial to auth-service failed")
				} else {
					log.Info().Caller().Str("address", s).Msg("dial to auth-service done")
				}
			}()
			return net.Dial("tcp", s)
		}),
	)
	if err != nil {
		log.Fatal().Caller().Str("auth", *auth).Err(err).Msg("dial failed")
		return
	}

	authClient := pbAuth.NewAuthClient(ccAuth)

	ccStorage, err := grpc.DialContext(ctx, *storage,
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

	storageClient := pbCRUD.NewCRUDClient(ccStorage)

	server := web.New(authClient, storageClient)

	server.Run(ctx, *port)
}
