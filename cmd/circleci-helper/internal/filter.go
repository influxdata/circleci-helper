package internal

import (
	"sort"
	"strings"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
)

// de-duplicate multiple workflows with same name, only picking up most recent workflow with the name
// this is required to allow retrying CircleCI workflows or jobs and only retrieving latest result
func uniqueWorkflows(workflows []*circle.Workflow) []*circle.Workflow {
	// sort workflows by creation time, descending, so that only most recent workflow with same name is used
	var sortedWorkflows []*circle.Workflow
	sortedWorkflows = append(sortedWorkflows, workflows...)
	sort.Slice(sortedWorkflows, func(a, b int) bool {
		return sortedWorkflows[a].CreatedAt > sortedWorkflows[b].CreatedAt
	})

	workflowAdded := map[string]bool{}

	var result []*circle.Workflow
	for _, workflow := range sortedWorkflows {
		if !workflowAdded[workflow.Name] {
			workflowAdded[workflow.Name] = true
			result = append(result, workflow)
		}
	}

	return result
}

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

func filterJobs(jobs []*circle.Job, excludeJobNames []string, jobPrefixes []string) []*circle.Job {
	if len(excludeJobNames) == 0 {
		return jobs
	}

	var result []*circle.Job
	for _, job := range jobs {
		matches := true
		if len(jobPrefixes) > 0 {
			// when limiting to prefixes, ensure at least one job has same prefix, otherwise ignore job
			matches = false
			for _, jobPrefix := range jobPrefixes {
				if strings.HasPrefix(job.Name, jobPrefix) {
					matches = true
					break
				}
			}
		}
		for _, excludeJobName := range excludeJobNames {
			if job.Name == excludeJobName {
				matches = false
				break
			}
		}
		if matches {
			result = append(result, job)
		}
	}
	return result
}
