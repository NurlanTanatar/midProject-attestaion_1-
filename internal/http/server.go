package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"midProject/internal/models"
	"midProject/internal/store"
	"midProject/tools"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Server struct {
	ctx         context.Context
	idleConnsCh chan struct{}
	store       store.Store

	Address string
}

func NewServer(ctx context.Context, address string, store store.Store) *Server {
	return &Server{
		ctx:         ctx,
		idleConnsCh: make(chan struct{}),
		store:       store,

		Address: address,
	}
}

func (s *Server) basicHandler() chi.Router {
	r := chi.NewRouter()

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()
		passwordB64 := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

		result := "denied"
		if access := tools.ValidateToken(passwordB64); access {
			result = "granted"
		}

		render.PlainText(w, r, result)
	})

	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		user := new(models.User)
		if err := json.NewDecoder(r.Body).Decode(user); err != nil {
			fmt.Fprintf(w, "Unknown err: %v", err)
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		user.ID = primitive.NewObjectID()
		user.ImgPath = tools.GenOTPREST(user)
		s.store.Create(r.Context(), user)
		w.WriteHeader(http.StatusCreated)
	})

	r.Post("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")

		user, err := s.store.ByID(r.Context(), idStr)
		if err != nil {
			fmt.Fprintf(w, "Unknown err: %v", err)
			return
		}

		type Token struct {
			Number string `json:"token"`
		}

		token := new(Token)

		if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
			fmt.Fprintf(w, "Unknown err: %v", err)
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		result := tools.GivePerm(user, token.Number)
		render.PlainText(w, r, result)
	})

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		users, err := s.store.All(r.Context())
		if err != nil {
			fmt.Fprintf(w, "Unknown err: %v", err)
			w.WriteHeader(http.StatusConflict)
			return
		}

		render.JSON(w, r, users)
	})
	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")

		user, err := s.store.ByID(r.Context(), idStr)
		if err != nil {
			fmt.Fprintf(w, "Unknown err: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		render.JSON(w, r, user)
	})

	r.Get("/users/{id}/qr", func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")

		user, err := s.store.ByID(r.Context(), idStr)
		if err != nil {
			fmt.Fprintf(w, "Unknown err: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		fileBytes, err := ioutil.ReadFile(user.ImgPath)
		if err != nil {
			fmt.Fprintf(w, "err: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(fileBytes)
	})

	r.Put("/users", func(w http.ResponseWriter, r *http.Request) {
		user := new(models.User)
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			fmt.Fprintf(w, "Unknown err: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		user.ImgPath = tools.GenOTPREST(user)
		err := s.store.Update(r.Context(), user)
		if err != nil {
			fmt.Fprintf(w, "err: %v", err)
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	r.Delete("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")

		if err := s.store.Delete(r.Context(), idStr); err != nil {
			fmt.Fprintf(w, "err: %v", err)
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return r
}

func (s *Server) Run() error {
	srv := &http.Server{
		Addr:         s.Address,
		Handler:      s.basicHandler(),
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 30,
	}
	go s.ListenCtxForGT(srv)

	log.Println("[HTTP] Server running on", s.Address)
	return srv.ListenAndServe()
}

func (s *Server) ListenCtxForGT(srv *http.Server) {
	<-s.ctx.Done() // блокируемся, пока контекст приложения не отменен

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("[HTTP] Got err while shutting down^ %v", err)
	}

	log.Println("[HTTP] Proccessed all idle connections")
	close(s.idleConnsCh)
}

func (s *Server) WaitForGracefulTermination() {
	// блок до записи или закрытия канала
	<-s.idleConnsCh
	os.RemoveAll("./tmp")
}
