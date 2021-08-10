package internal

import (
	"context"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
)

// WorkflowDetails provides information on CircleCI workflows that have not yet finished with details on individual job statuses.
type WorkflowDetails struct {
	Workflow      *circle.Workflow
	Failed        bool
	SucceededJobs []*circle.Job
	FailedJobs    []*circle.Job
	PendingJobs   []*circle.Job
}

// WorkflowsSummary provides summary on all workflows matching pattern and groups them into categories for easier reporting.
type WorkflowsSummary struct {
	Failed             bool
	Finished           bool
	AllWorkflows       []*WorkflowDetails
	SucceededWorkflows []*WorkflowDetails
	FailedWorkflows    []*WorkflowDetails
	PendingWorkflows   []*WorkflowDetails
}

// prepareWorkflowDetails prepares WorkflowDetails for a workflow, listing and filtering jobs and grouping them by status.
func prepareWorkflowDetails(
	ctx context.Context,
	client circle.Client,
	workflow *circle.Workflow,
	filterJob func(job *circle.Job) bool,
	listJobs bool,
) (*WorkflowDetails, error) {
	workflowDetails := &WorkflowDetails{
		Workflow: workflow,
		Failed:   circle.WorkflowFailed(workflow),
	}

	if listJobs {
		jobs, err := client.GetWorkflowJobs(ctx, workflow.ID)
		if err != nil {
			return nil, err
		}

		for _, job := range jobs {
			// if the job should not be considered, continue
			if !filterJob(job) {
				continue
			}

			if !circle.JobFinished(job) {
				// if the job has not finished yet, store it in the list of pending jobs
				workflowDetails.PendingJobs = append(workflowDetails.PendingJobs, job)
			} else if circle.JobFailed(job) {
				// if the job has failed, store it as a failed job
				workflowDetails.FailedJobs = append(workflowDetails.FailedJobs, job)
			} else {
				// if the job has finished, store it as a successful job
				workflowDetails.SucceededJobs = append(workflowDetails.SucceededJobs, job)
			}
		}
	}

	return workflowDetails, nil
}

// getLatestWorkflows returns latest workflows from specific pipelineID, filtered by callback function.
func getLatestWorkflows(
	ctx context.Context,
	client circle.Client,
	pipelineID string,
	filterWorkflow func(workflow *circle.Workflow) bool,
) ([]*circle.Workflow, error) {
	allWorkflows, err := client.GetWorkflows(ctx, pipelineID)
	if err != nil {
		return nil, err
	}

	workflows := []*circle.Workflow{}
	for _, workflow := range uniqueWorkflows(allWorkflows) {
		if filterWorkflow(workflow) {
			workflows = append(workflows, workflow)
		}
	}

	return workflows, nil
}

// checkWorkflowsStatus generates details for all workflows, filtering workflows and jobs, optionally also retrieving details for all or specific types of workflows / jobs
func checkWorkflowsStatus(
	ctx context.Context,
	client circle.Client,
	pipelineID string,
	filterWorkflow func(workflow *circle.Workflow) bool,
	filterJob func(job *circle.Job) bool,
	succeededJobDetails bool,
	failedJobDetails bool,
	pendingJobDetails bool,
) (*WorkflowsSummary, error) {
	result := &WorkflowsSummary{}

	workflows, err := getLatestWorkflows(ctx, client, pipelineID, filterWorkflow)
	if err != nil {
		return nil, err
	}

	result.Finished = true
	for _, workflow := range workflows {
		if circle.WorkflowFinished(workflow) {
			// if the workflow has finished, store it either as successful or failed
			if circle.WorkflowFailed(workflow) {
				workflowDetails, err := prepareWorkflowDetails(ctx, client, workflow, filterJob, failedJobDetails)
				if err != nil {
					return nil, err
				}

				result.FailedWorkflows = append(result.FailedWorkflows, workflowDetails)
				result.AllWorkflows = append(result.AllWorkflows, workflowDetails)

				result.Failed = true
			} else {
				workflowDetails, err := prepareWorkflowDetails(ctx, client, workflow, filterJob, succeededJobDetails)
				if err != nil {
					return nil, err
				}

				result.SucceededWorkflows = append(result.SucceededWorkflows, workflowDetails)
				result.AllWorkflows = append(result.AllWorkflows, workflowDetails)

			}
			// continue as the current workflow has already been added
			continue
		}

		workflowDetails, err := prepareWorkflowDetails(ctx, client, workflow, filterJob, pendingJobDetails)
		if err != nil {
			return nil, err
		}

		if pendingJobDetails {
			// if we've retrieved pending job details, assume the workflow has failed if at least one job has already failed and function was asked to retrieve details
			if len(workflowDetails.PendingJobs) > 0 {
				result.Finished = false
			}
		} else {
			// if we don't have details on pending jobs, assume a pending workflow means not everything has finished
			result.Finished = false
		}

		// assume the workflow has failed if at least one job has already failed and function was asked to retrieve details
		if len(workflowDetails.FailedJobs) > 0 {
			result.Failed = true
		}

		result.PendingWorkflows = append(result.PendingWorkflows, workflowDetails)
		result.AllWorkflows = append(result.AllWorkflows, workflowDetails)
	}

	return result, nil
}
