package handlers

import (
	"encoding/json"
	"fmt"
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
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "job not found",
		})
		return
	}

	json.NewEncoder(w).Encode(j)
}

func (a *Application) EnqueueJob(w http.ResponseWriter, r *http.Request) {
	var j queue.Job

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&j); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "invalid job",
		})
		return
	}

	if j.Type != queue.JobType_TimeCritical && j.Type != queue.JobType_NotTimeCritical {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "invalid job type",
		})
		return
	}

	a.Queue.Enqueue(&j)

	json.NewEncoder(w).Encode(j)
}

func (a *Application) DequeueJob(w http.ResponseWriter, r *http.Request) {
	j, err := a.Queue.Dequeue()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("could not dequeue job: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(j)
}

func (a *Application) ConcludeJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	j, err := a.Queue.Conclude(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("could not conclude job: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(j)
}
