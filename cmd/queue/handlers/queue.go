package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jscottmiller/job-queue/internal/queue"
)

func (a *Application) ReadJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	j := a.Queue.GetJob(id)
	if j == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(j)
}

func (a *Application) EnqueueJob(w http.ResponseWriter, r *http.Request) {
	var j queue.Job

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&j); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a.Queue.Enqueue(&j)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(j)
}

func (a *Application) DequeueJob(w http.ResponseWriter, r *http.Request) {
	j, err := a.Queue.Dequeue()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(j)
}

func (a *Application) ConcludeJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := a.Queue.Conclude(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}
