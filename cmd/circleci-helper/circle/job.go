package circle

// Job describes a single CircleCI job
type Job struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// JobFinished returns whether specified job has finished and is no longer in progress
func JobFinished(job *Job) bool {
	return JobFailed(job) || job.Status == "success"
}

// JobFailed returns whether specified job has failed
func JobFailed(job *Job) bool {
	return job.Status == "failed"
}
