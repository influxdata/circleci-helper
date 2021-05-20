package internal

import (
	"context"
	"time"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
	"go.uber.org/zap"
)

// WorkflowsSummary provides summary on all workflows matching pattern and groups them into categories for easier reporting.
type WorkflowsSummary struct {
	Finished           bool
	SucceededWorkflows []*circle.Workflow
	FailedWorkflows    []*circle.Workflow
	PendingWorkflows   []*PendingWorkflowDetails
}

// PendingWorkflowDetails provides information on CircleCI workflows that have not yet finished with details on individual job statuses.
type PendingWorkflowDetails struct {
	Workflow      *circle.Workflow
	SucceededJobs []*circle.Job
	FailedJobs    []*circle.Job
	PendingJobs   []*circle.Job
}

func checkWorkflowsStatus(ctx context.Context, client circle.Client, pipelineID string, workflowNames []string, excludeJobNames []string) (*WorkflowsSummary, error) {
	result := &WorkflowsSummary{}

	workflows, err := client.GetWorkflows(ctx, pipelineID)
	if err != nil {
		return result, err
	}

	workflows = filterWorkflows(uniqueWorkflows(workflows), workflowNames)

	// assume finished is true if workflows matched unless at least one of them is still pending
	// if not all of the reported workflows were returned by filters, assume it is not finished and return
	if len(workflows) < len(workflowNames) {
		return result, nil
	}

	result.Finished = true
	for _, workflow := range workflows {
		if circle.WorkflowFinished(workflow) {
			// if the workflow has finished, store it either as successful or failed
			if circle.WorkflowFailed(workflow) {
				result.FailedWorkflows = append(result.FailedWorkflows, workflow)
			} else {
				result.SucceededWorkflows = append(result.SucceededWorkflows, workflow)
			}
			// continue as the current workflow has already been added
			continue
		}

		// analyze specific jobs, excluding jobs requested by the called
		jobs, err := client.GetWorkflowJobs(ctx, workflow.ID)
		if err != nil {
			return result, err
		}

		jobs = filterJobs(jobs, excludeJobNames)

		// store the workflow and details about each job in the result
		pendingWorkflow := &PendingWorkflowDetails{Workflow: workflow}
		result.PendingWorkflows = append(result.PendingWorkflows, pendingWorkflow)

		for _, job := range jobs {
			if circle.JobFinished(job) {
				// if the job has finished, store it either as successful or failed
				if circle.JobFailed(job) {
					pendingWorkflow.FailedJobs = append(pendingWorkflow.FailedJobs, job)
				} else {
					pendingWorkflow.SucceededJobs = append(pendingWorkflow.SucceededJobs, job)
				}
			} else {
				// if the job has not finished yet, store it in the list of pending jobs
				pendingWorkflow.PendingJobs = append(pendingWorkflow.PendingJobs, job)

				// if one or more jobs inside non-finished workflows have not finished, assume the pipeline has not finished
				result.Finished = false
			}
		}
	}

	return result, nil
}

// WaitForJobs waits for all jobs matching criteria to finish, ignoring their results.
func WaitForJobs(ctx context.Context, logger *zap.Logger, client circle.Client, projectType string, org string, project string, pipelineNumber int, workflowNames []string, excludeJobNames []string) (*WorkflowsSummary, error) {
	sugar := logger.Sugar()

	pipelineID, err := client.GetPipelineID(ctx, projectType, org, project, pipelineNumber)
	if err != nil {
		return nil, err
	}

	// loop forever, timeout is handled by the context ; any API requests to CircleCI
	// after timeout will fail and the loop will exit with an error
	for {
		result, err := checkWorkflowsStatus(ctx, client, pipelineID, workflowNames, excludeJobNames)
		if err != nil {
			return nil, err
		}

		// report all workflows - starting with
		for _, workflow := range result.SucceededWorkflows {
			sugar.Infof("workflow %s finished (status: %s)", workflow.Name, workflow.Status)
		}

		for _, workflow := range result.FailedWorkflows {
			sugar.Warnf("workflow %s failed (status: %s)", workflow.Name, workflow.Status)
		}

		for _, details := range result.PendingWorkflows {
			sugar.Infof("workflow %s has not finished yet (status: %s)", details.Workflow.Name, details.Workflow.Status)
			for _, job := range details.SucceededJobs {
				sugar.Infof("  - job %s finished (status: %s)", job.Name, job.Status)
			}
			for _, job := range details.PendingJobs {
				sugar.Infof("  - job %s in progress (status: %s)", job.Name, job.Status)
			}
			for _, job := range details.FailedJobs {
				sugar.Warnf("  - job %s failed (status: %s)", job.Name, job.Status)
			}
		}

		// if everything has finished already, simply report that and return
		if result.Finished {
			sugar.Infof("all workflows finished successfully")
			return result, nil
		}

		// if one more workflows have not finished, wait and try again
		sugar.Infof("Not all workflows / jobs have finished, waiting\n")
		time.Sleep(30 * time.Second)
	}
}
