package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jscottmiller/job-queue/internal/queue"
)

type Application struct {
	Queue *queue.Queue
}

type ErrorResponse struct {
	Error string
}

func (a *Application) Router() *mux.Router {
	r := mux.NewRouter()
	r.Use(jsonMiddleware)

	jobs := r.PathPrefix("/jobs").Subrouter()
	jobs.HandleFunc("/enqueue", a.EnqueueJob).Methods("POST")
	jobs.HandleFunc("/dequeue", a.DequeueJob).Methods("POST")
	jobs.HandleFunc("/{id:[a-z0-9\\-]+}", a.ReadJob).Methods("GET")
	jobs.HandleFunc("/{id:[a-z0-9\\-]+}/conclude", a.ConcludeJob).Methods("POST")

	return r
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
