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

func filterWorkflow(workflow *circle.Workflow, keepNames []string) bool {
	if len(keepNames) == 0 {
		return true
	}

	for _, workflowName := range keepNames {
		if workflow.Name == workflowName {
			return true
		}
	}

	return false
}

func filterWorkflows(workflows []*circle.Workflow, keepNames []string) []*circle.Workflow {
	var result []*circle.Workflow
	for _, workflow := range workflows {
		if filterWorkflow(workflow, keepNames) {
			result = append(result, workflow)
		}
	}
	return result
}

func filterJob(job *circle.Job, excludeJobNames []string, jobPrefixes []string) bool {
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
	return matches
}

func filterWorkflowWrapper(keepNames []string) func(workflow *circle.Workflow) bool {
	return func(workflow *circle.Workflow) bool {
		return filterWorkflow(workflow, keepNames)
	}
}

func filterJobWrapper(excludeJobNames []string, jobPrefixes []string) func(job *circle.Job) bool {
	return func(job *circle.Job) bool {
		return filterJob(job, excludeJobNames, jobPrefixes)
	}
}
