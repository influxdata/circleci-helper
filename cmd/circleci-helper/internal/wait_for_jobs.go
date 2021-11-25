package internal

import (
	"context"
	"math"
	"time"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
	"go.uber.org/zap"
)

// WaitForJobsOptions allows passing options for retrieving status of one or more workflows.
type WaitForJobsOptions struct {
	ProjectType              string
	Org                      string
	Project                  string
	PipelineNumber           int
	WorkflowNames            []string
	ExcludeJobNames          []string
	JobPrefixes              []string
	FailOnError              bool
	GetSucceededWorkflowJobs bool
	GetFailedWorkflowJobs    bool
	GetPendingWorkflowJobs   bool
	WaitDuration             *WaitForJobsDuration
}

// WaitForJobs waits for all jobs matching criteria to finish, ignoring their results.
func WaitForJobs(ctx context.Context, logger *zap.Logger, client circle.Client, opts WaitForJobsOptions) (*WorkflowsSummary, error) {
	sugar := logger.Sugar()

	pipelineID, err := client.GetPipelineID(ctx, opts.ProjectType, opts.Org, opts.Project, opts.PipelineNumber)
	if err != nil {
		return nil, err
	}

	// loop forever, timeout is handled by the context ; any API requests to CircleCI
	// after timeout will fail and the loop will exit with an error
	var pendingJobCount int

	for {
		result, err := checkWorkflowsStatus(
			ctx, client, pipelineID,
			checkWorkflowStatusOpts{
				filterWorkflow:    filterWorkflowWrapper(opts.WorkflowNames),
				filterJob:         filterJobWrapper(opts.ExcludeJobNames, opts.JobPrefixes),
				pendingJobDetails: true,
			},
		)

		if err != nil {
			return nil, err
		}

		// assume finished is true if workflows matched unless at least one of them is still pending
		// if not all of the reported workflows were returned by filters, assume it is not finished and use an empty result
		if len(result.AllWorkflows) < len(opts.WorkflowNames) {
			result = &WorkflowsSummary{}
		}

		pendingJobCount = 0

		// report all workflows - starting with successful ones
		for _, workflowDetails := range result.SucceededWorkflows {
			sugar.Infof("workflow %s finished (status: %s)", workflowDetails.Workflow.Name, workflowDetails.Workflow.Status)
		}

		for _, workflow := range result.FailedWorkflows {
			sugar.Warnf("workflow %s failed (status: %s)", workflow.Workflow.Name, workflow.Workflow.Status)
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
			pendingJobCount += len(details.PendingJobs)
		}

		// if everything has finished already, simply report that and return
		if result.Finished {
			if result.Failed {
				sugar.Warnf("all workflows finished - failed")
			} else {
				sugar.Infof("all workflows finished - successfully")
			}
			return result, nil
		}

		if result.Failed && opts.FailOnError {
			sugar.Warnf("one or workflows has failed and should fail on error - exiting")
			return result, nil
		}

		// if one more workflows have not finished, wait and try again
		duration := opts.WaitDuration.GetDuration(pendingJobCount)
		sugar.Infof("Not all workflows / jobs have finished, waiting for %g seconds", math.Round(duration.Seconds()))
		time.Sleep(duration)
	}
}
