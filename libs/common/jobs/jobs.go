package jobs

import (
	"log"
	"runtime/debug"
	"time"

	"github.com/go-co-op/gocron"
)

type JobFunc func() error
type jobInfo struct {
	name string
	job  JobFunc
}
type JobManager struct {
	scheduler *gocron.Scheduler
	jobs      map[string]jobInfo
}

type Interval uint16

const (
	IntervalHourly   = 1
	IntervalMinutely = 2
)

func NewJobManager() *JobManager {
	scheduler := gocron.NewScheduler(time.UTC)

	return &JobManager{
		scheduler: scheduler,
		jobs:      make(map[string]jobInfo),
	}
}

func (j *JobManager) Start() {
	j.scheduler.StartBlocking()
}

func (j *JobManager) Stop() {
	j.scheduler.Stop()
}

func (j *JobManager) Cron(cronStr string, name string, job JobFunc, enabled bool) {
	jobInfo := jobInfo{
		name: name,
		job:  job,
	}
	j.jobs[name] = jobInfo
	if enabled {
		j.scheduler.Cron(cronStr).Do(func() {
			j.RunJob(name)
		})
	}
}

func (j *JobManager) RunJob(name string) {
	// TODO: distributed locks with redis using https://github.com/go-redsync/redsync
	job, ok := j.jobs[name]
	if !ok {
		return
	}
	start := time.Now()
	log.Printf("Starting job %q", job.name)
	defer func(job jobInfo) {
		if r := recover(); r != nil {
			log.Printf("Job %q panicked: %v \n stacktrace: %v", job.name, r, string(debug.Stack()))
		}
	}(job)
	err := job.job()
	if err != nil {
		duration := time.Since(start)
		log.Printf("Job %q failed after %v ms: %v", job.name, duration.Milliseconds(), err)
	} else {
		duration := time.Since(start)
		log.Printf("Job %q completed successfully after %v ms", job.name, duration.Milliseconds())
	}

}
