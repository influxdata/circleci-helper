package internal

import (
	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
)

func filterWorkflows(workflows []*circle.Workflow, keepNames []string) []*circle.Workflow {
	if len(keepNames) == 0 {
		return workflows
	}

	var result []*circle.Workflow
	for _, workflow := range workflows {
		matches := false
		for _, workflowName := range keepNames {
			if workflow.Name == workflowName {
				matches = true
				break
			}
		}
		if matches {
			result = append(result, workflow)
		}
	}
	return result
}

func filterJobs(jobs []*circle.Job, excludeJobNames []string) []*circle.Job {
	if len(excludeJobNames) == 0 {
		return jobs
	}

	var result []*circle.Job
	for _, job := range jobs {
		exclude := false
		for _, excludeJobName := range excludeJobNames {
			if job.Name == excludeJobName {
				exclude = true
				break
			}
		}
		if !exclude {
			result = append(result, job)
		}
	}
	return result
}
