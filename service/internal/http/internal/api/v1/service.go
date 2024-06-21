package v1

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Runner is the underlying data processing layer of Service.
type Runner interface {
	// ExecJob should schedule the job to run asynchronously and return with the current running jobID.
	ExecJob(ctx context.Context, content []byte) (jobID string, err error)
	// GetJobstatus is used to get the current status of a job or return exist==false if the job does not exist.
	GetJobStatus(jobID string) (status string, exist bool)
	// ReadJobPerformance is used to get the running performance for a given jobID.
	ReadJobPerformance(jobID string) (float64, error)
}

// Service defines all routes handled by Service
type Service struct {
	runner Runner
}

func NewService(runner Runner) *Service {
	s := &Service{
		runner: runner,
	}
	return s
}

func (s *Service) Run(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobID, err := s.runner.ExecJob(ctx, content)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Write([]byte(fmt.Sprintf("Job was launched with id %v\n", jobID)))
}

func (s *Service) Status(w http.ResponseWriter, r *http.Request) {
	// Get request query parameters
	vars := mux.Vars(r)
	jobID := vars["jobid"]

	// check status
	status, ok := s.runner.GetJobStatus(jobID)
	if !ok {
		http.Error(w, fmt.Sprintf("Job with id %s was not found", jobID), http.StatusNotFound)
		return
	}

	perf, err := s.runner.ReadJobPerformance(jobID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("Job %s was %s - Model performance : %f", jobID, status, perf)))
}
