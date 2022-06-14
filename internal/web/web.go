package web

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strconv"

	pbAuth "github.com/amasynikov/grpc-webinar/internal/genproto/auth"
	pbCRUD "github.com/amasynikov/grpc-webinar/internal/genproto/crud"
)

type ctxIkKey struct{}

type httpSever struct {
	auth    pbAuth.AuthClient
	storage pbCRUD.CRUDClient
}

func New(
	auth pbAuth.AuthClient,
	storage pbCRUD.CRUDClient,
) *httpSever {
	return &httpSever{
		auth:    auth,
		storage: storage,
	}
}

func (s httpSever) Run(ctx context.Context, port int) {
	root := mux.NewRouter()

	root.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Trace().Caller().Str("method", r.Method).Str("path", r.RequestURI).Msg("")
			next.ServeHTTP(w, r)
		})
	})

	root.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("user")
		password, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		loginOk, err := s.auth.Login(r.Context(), &pbAuth.LoginRequest{
			User:     user,
			Password: string(password),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(loginOk.GetToken()))
	})).Methods(http.MethodPost)

	routes := root.NewRoute().Subrouter()

	routes.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Header.Get("user")
			token := r.Header.Get("token")
			_, err := s.auth.Validate(r.Context(), &pbAuth.ValidateRequest{
				User:  user,
				Token: token,
			})
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			}
			id := mux.Vars(r)["id"]
			next.ServeHTTP(
				w,
				r.WithContext(
					context.WithValue(
						r.Context(),
						ctxIkKey{},
						id,
					),
				),
			)
		})
	})

	routes.Handle("/create", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}
		createOk, err := s.storage.Create(request.Context(), &pbCRUD.CreateRequest{
			Raw: body,
		})
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(createOk.GetId()))
	})).Methods(http.MethodPut)

	routes.Handle("/read/{id}", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		id := request.Context().Value(ctxIkKey{}).(string)
		readOk, err := s.storage.Read(request.Context(), &pbCRUD.ReadRequest{Id: id})
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write(readOk.GetRaw())
	})).Methods(http.MethodGet)

	routes.Handle("/update/{id}", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		id := request.Context().Value(ctxIkKey{}).(string)
		body, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}
		_, err = s.storage.Update(request.Context(), &pbCRUD.UpdateRequest{
			Data: &pbCRUD.Data{
				Id:  id,
				Raw: body,
			},
		})
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}
		writer.WriteHeader(http.StatusOK)
	})).Methods(http.MethodPost)

	routes.Handle("/delete/{id}", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		id := request.Context().Value(ctxIkKey{}).(string)
		_, err := s.storage.Delete(request.Context(), &pbCRUD.DeleteRequest{
			Id: id,
		})
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}
		writer.WriteHeader(http.StatusOK)
	})).Methods(http.MethodDelete)

	if err := http.ListenAndServe(":"+strconv.Itoa(port), root); err != nil {
		log.Fatal().Caller().Err(err).Msg("")
	}
}
