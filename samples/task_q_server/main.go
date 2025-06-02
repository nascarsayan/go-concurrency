package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"
)

type JobReq struct {
	Job    Job
	result chan Job
}

type Job struct {
	Id           int
	DurationSecs int
	SubmittedAt  *time.Time
	StartedAt    *time.Time
	CompletedAt  *time.Time
}

type StatusReq struct {
	report chan string
}

var queue *Queue

type Queue struct {
	// Job to jobReqCh should be sent here.
	jobReqCh chan JobReq
	// Status request should be sent here.
	statusReqCh chan StatusReq
	// A job just completed should be sent back here.
	doneCh    chan *JobReq
	Running   []*JobReq
	Pending   []*JobReq
	nextJobId int
}

func (q *Queue) Submit(job Job) chan Job {
	if job.Id == 0 {
		q.nextJobId++
		job.Id = q.nextJobId
	}
	now := time.Now().Local()
	job.SubmittedAt = &now
	req := JobReq{
		Job:    job,
		result: make(chan Job),
	}
	q.jobReqCh <- req
	fmt.Printf("Job submitted at %s\n", now.Format(time.RFC3339))
	return req.result
}

func (q *Queue) Status() chan string {
	req := StatusReq{
		report: make(chan string),
	}
	q.statusReqCh <- req
	return req.report
}

func (jr *JobReq) Start(doneCh chan *JobReq) {
	now := time.Now().Local()
	fmt.Printf("Job started at %s\n", now.Format(time.RFC3339))
	jr.Job.StartedAt = &now
	time.Sleep(time.Second * time.Duration(jr.Job.DurationSecs))
	now = time.Now().Local()
	jr.Job.CompletedAt = &now
	fmt.Printf("Job completed at %s\n", now.Format(time.RFC3339))
	jr.result <- jr.Job
	close(jr.result)
	doneCh <- jr
}

const MAX_PARALLEL = 4

func (q *Queue) getStatus() string {
	b, err := json.MarshalIndent(q, "", "  ")
	fmt.Printf("Queue status: %s\n", string(b))
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(b)
}

func (q *Queue) Run() {
	for {
		select {
		case req := <-q.jobReqCh:
			fmt.Printf("Job submitted at %s\n", req.Job.SubmittedAt.Format(time.RFC3339))
			if len(q.Running) == 4 {
				fmt.Printf("All cores occupied, sending to pending")
				q.Pending = append(q.Pending, &req)
				continue
			}
			q.nextJobId++
			if req.Job.Id == 0 {
				req.Job.Id = q.nextJobId
			}
			q.Running = append(q.Running, &req)
			go req.Start(q.doneCh)
		case done := <-q.doneCh:
			// remove job id from the running queue
			for i := range len(q.Running) {
				if q.Running[i].Job.Id == done.Job.Id {
					q.Running = slices.Delete(q.Running, i, i+1)
					break
				}
			}
			if len(q.Pending) == 0 || len(q.Running) == MAX_PARALLEL {
				continue
			}
			// if there are pending jobs, start the next one
			for len(q.Pending) > 0 && len(q.Running) < MAX_PARALLEL {
				req := q.Pending[0]
				q.Pending = q.Pending[1:]
				q.Running = append(q.Running, req)
				go req.Start(q.doneCh)
			}
		case req := <-q.statusReqCh:
			fmt.Printf("Status request received at %s\n", time.Now().Format(time.RFC3339))
			status := q.getStatus()
			req.report <- status
			close(req.report)
		}
	}
}

func initQueue() *Queue {
	q := new(Queue)
	q.jobReqCh = make(chan JobReq)
	q.statusReqCh = make(chan StatusReq)
	q.doneCh = make(chan *JobReq)
	q.Running = make([]*JobReq, 0)
	q.Pending = make([]*JobReq, 0)
	go q.Run()
	return q
}

func handleJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Time int `json:"time"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	now := time.Now().Local()
	j := Job{
		DurationSecs: data.Time,
		SubmittedAt:  &now,
	}
	if j.DurationSecs <= 0 {
		http.Error(w, "Invalid duration", http.StatusBadRequest)
		return
	}
	jobCh := queue.Submit(j)
	output := <-jobCh
	b, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		http.Error(w, "Json Marshal failed", http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "%s", b)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	statusCh := queue.Status()
	status := <-statusCh
	fmt.Fprintf(w, "%s", bytes.NewBufferString(status))
}

func main() {
	queue = initQueue()
	http.HandleFunc("/job", handleJob)
	http.HandleFunc("/status", handleStatus)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
