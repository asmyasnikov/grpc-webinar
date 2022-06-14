package auth

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"sync"
	"time"

	grpc_codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/amasynikov/grpc-webinar/internal/genproto/auth"
)

var (
	errUnauthenticated  = status.Errorf(grpc_codes.Unauthenticated, "wrong login+password pair")
	errWrongCredentials = status.Errorf(grpc_codes.Unauthenticated, "wrong login+token pair")
)

type token struct {
	user    string
	created time.Time
}

type authServer struct {
	pb.UnimplementedAuthServer

	// read-only success
	validUserPasswords map[string]string

	// read-write success
	tokensMtx sync.RWMutex
	tokens    map[string]token
}

func (a *authServer) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	t, err := a.login(ctx, request.GetUser(), request.GetPassword())
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{Token: t}, nil
}

func (a *authServer) login(ctx context.Context, login, password string) (t string, err error) {
	log.Trace().Caller().Str("user", login).Msg("login")
	defer func() {
		if err != nil {
			log.Warn().Caller().Str("user", login).Err(err).Msg("login failed")
		} else {
			log.Info().Caller().Str("user", login).Msg("login successfully")
		}
	}()

	if p, ok := a.validUserPasswords[login]; !ok || p != password {
		return "", errUnauthenticated
	}

	tt, err := uuid.NewUUID()
	if err != nil {
		return "", status.Errorf(grpc_codes.Internal, err.Error())
	}

	a.tokensMtx.Lock()
	defer a.tokensMtx.Unlock()

	a.tokens[tt.String()] = token{
		user:    login,
		created: time.Now(),
	}

	return tt.String(), nil
}

func (a *authServer) Validate(ctx context.Context, request *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	err := a.validate(ctx, request.GetUser(), request.GetToken())
	if err != nil {
		return nil, err
	}
	return &pb.ValidateResponse{}, nil
}

func (a *authServer) validate(ctx context.Context, login, token string) (err error) {
	log.Trace().Caller().Str("user", login).Msg("validate")
	defer func() {
		if err != nil {
			log.Warn().Caller().Str("user", login).Err(err).Msg("validate failed")
		} else {
			log.Debug().Caller().Str("user", login).Msg("validate successfully")
		}
	}()

	a.tokensMtx.RLock()
	defer a.tokensMtx.RUnlock()

	if t, ok := a.tokens[token]; ok && t.user == login {
		return nil
	}

	return errWrongCredentials
}

func (a *authServer) clearTokens(ttl time.Duration) {
	log.Trace().Caller().Dur("ttl", ttl).Msg("clearTokens start")
	defer func() {
		log.Trace().Caller().Dur("ttl", ttl).Msg("clearTokens done")
	}()

	a.tokensMtx.Lock()
	defer a.tokensMtx.Unlock()

	var tokensToDelete []string

	for t, tt := range a.tokens {
		if time.Since(tt.created) > ttl {
			log.Debug().Caller().Str("token", t).Msg("will be deleted")
			tokensToDelete = append(tokensToDelete, t)
		}
	}

	for _, t := range tokensToDelete {
		delete(a.tokens, t)
	}
}

func New(ttl time.Duration) *authServer {
	a := &authServer{
		validUserPasswords: map[string]string{
			"James":    "Holden",
			"Julia":    "Roberts",
			"Antonio":  "Banderos",
			"Olga":     "Buzova",
			"Courtney": "Love",
		},
		tokens: make(map[string]token),
	}
	if ttl > 0 {
		go func() {
			ticker := time.NewTicker(ttl)
			for range ticker.C {
				a.clearTokens(ttl)
			}
		}()
	}
	return a
}
