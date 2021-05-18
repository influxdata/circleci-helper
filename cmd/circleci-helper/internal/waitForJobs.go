package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
)

// WaitForJobs waits for all jobs matching criteria to finish, ignoring their results.
func WaitForJobs(token string, projectType string, org string, project string, pipelineNumber int, workflowNames []string, excludeJobNames []string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	pipelineID, err := circle.GetPipelineID(ctx, token, projectType, org, project, pipelineNumber)
	if err != nil {
		return false, err
	}

	for {
		workflows, err := circle.GetWorkflows(ctx, token, pipelineID)
		if err != nil {
			return false, err
		}

		workflows = filterWorkflows(workflows, workflowNames)

		allJobsFinished := false
		success := true

		// only compare results if all of the expected workflows were reported
		if len(workflows) >= len(workflowNames) {
			allJobsFinished = true
			for _, workflow := range workflows {
				// if entire workflow has not finished yet, analyze specific jobs, excluding jobs we do not care about
				if circle.WorkflowFinished(workflow) {
					if circle.WorkflowFailed(workflow) {
						success = false
						fmt.Printf("FAIL Workflow %s failed\n", workflow.Name)
					} else {
						fmt.Printf("OK   Workflow %s succeeded\n", workflow.Name)
					}
				} else {
					jobs, err := circle.GetWorkflowJobs(ctx, token, workflow.ID)
					if err != nil {
						return false, err
					}

					jobs = filterJobs(jobs, excludeJobNames)

					for _, job := range jobs {
						if circle.JobFinished(job) {
							if circle.JobFailed(job) {
								success = false
								fmt.Printf("FAIL Workflow %s job %s failed (status: %s)\n", workflow.Name, job.Name, job.Status)
							} else {
								fmt.Printf("OK   Workflow %s job %s finished (status: %s)\n", workflow.Name, job.Name, job.Status)
							}
						} else {
							fmt.Printf("WARN Workflow %s job %s not yet finished (status: %s)\n", workflow.Name, job.Name, job.Status)
							allJobsFinished = false
						}
					}
				}
			}
		}

		if allJobsFinished {
			return success, nil
		}

		fmt.Printf("WARN Not all workflows / jobs have finished, waiting\n")
		time.Sleep(30 * time.Second)
	}

	// reasonable default if the return was never reached
	return false, nil
}
