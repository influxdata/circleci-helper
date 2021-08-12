package internal

import (
	"context"
	"strings"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
	"go.uber.org/zap"
)

// WaitForJobsOptions allows passing options for retrieving status of one or more workflows.
type WorkflowErrorsOptions struct {
	ProjectType    string
	Org            string
	Project        string
	PipelineNumber int
	WorkflowNames  []string
}

type WorkflowErrorsFailure struct {
	Workflow   *circle.Workflow
	Job        *circle.Job
	StepName   string
	ActionName string
	Messages   string
}

type WorkflowErrorsResult struct {
	Failures []*WorkflowErrorsFailure
}

// WorkflowErrors retrieves all errors for a workflow
func WorkflowErrors(ctx context.Context, logger *zap.Logger, client circle.Client, opts WorkflowErrorsOptions) (*WorkflowErrorsResult, error) {
	pipelineID, err := client.GetPipelineID(ctx, opts.ProjectType, opts.Org, opts.Project, opts.PipelineNumber)
	if err != nil {
		return nil, err
	}

	status, err := checkWorkflowsStatus(
		ctx, client, pipelineID,
		checkWorkflowStatusOpts{
			filterWorkflow: filterWorkflowWrapper(opts.WorkflowNames),
			// retrieve details for all types of jobs
			succeededJobDetails: true,
			failedJobDetails:    true,
			pendingJobDetails:   true,
		},
	)

	if err != nil {
		return nil, err
	}

	// assume finished is true if workflows matched unless at least one of them is still pending
	// if not all of the reported workflows were returned by filters, assume it is not finished and use an empty result
	if len(status.AllWorkflows) < len(opts.WorkflowNames) {
		status = &WorkflowsSummary{}
	}

	result := &WorkflowErrorsResult{
		Failures: []*WorkflowErrorsFailure{},
	}

	for _, workflow := range status.AllWorkflows {
		jobs := append(workflow.FailedJobs, workflow.PendingJobs...)

		for _, job := range jobs {
			// ignore jobs that were blocked by other dependencies since they do not have any details to retrieve
			if job.Status == "blocked" {
				continue
			}

			details, err := client.GetJobDetails(ctx, opts.ProjectType, opts.Org, opts.Project, job.JobNumber)
			if err != nil {
				// check if the error was 404 - if so, assume the job has not yet been run and continue
				httpErr, ok := err.(*circle.ClientHTTPError)
				if ok && httpErr.StatusCode == 404 {
					continue
				}

				// if this was not a 404 error, return the real error
				return nil, err
			}

			for _, step := range details.Steps {
				for _, action := range step.Actions {
					if action.Failed {
						output, err := client.GetJobActionOutput(ctx, &action)
						if err != nil {
							return nil, err
						}

						var sb strings.Builder
						for _, line := range output {
							sb.WriteString(line.Message)
						}

						result.Failures = append(result.Failures, &WorkflowErrorsFailure{
							Workflow:   workflow.Workflow,
							Job:        job,
							StepName:   step.Name,
							ActionName: action.Name,
							Messages:   sb.String(),
						})
					}
				}
			}
		}
	}

	return result, nil
}
