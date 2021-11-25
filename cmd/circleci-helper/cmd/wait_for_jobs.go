package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
	"github.com/influxdata/circleci-helper/cmd/circleci-helper/internal"
)

var exclude string
var jobPrefix string
var failOnError bool
var failHeader string
var failFooter string
var timeout time.Duration
var waitTime time.Duration

// waitForJobsCmd represents the waitForJobs command
var waitForJobsCmd = &cobra.Command{
	Use:   "wait-for-jobs",
	Short: "Wait for other job or jobs in specified workflow",
	Long: `Wait for one or more other jobs in specified workflow. For example:

circleci-helper wait-for-jobs --token ... --pipeline ... --workflow "myworkflow" --project-type ... --exclude "my-finalize-job"
`,
	Run: func(cmd *cobra.Command, args []string) {
		commandHelper(cmd, args, waitForJobsMain)
	},
}

func printWorkflowNameAndURL(workflow *circle.Workflow) {
	workflowURL := fmt.Sprintf(
		"https://app.circleci.com/pipelines/%s/%s/%s/%d/workflows/%s",
		url.PathEscape(projectType), url.PathEscape(org), url.PathEscape(project),
		pipelineNumber,
		url.PathEscape(workflow.ID),
	)
	fmt.Printf("  - %s ( %s )\n", workflow.Name, workflowURL)
}

func waitForJobsMain(logger *zap.Logger, cmd *cobra.Command, args []string) error {
	if err := validateWorkflowFlags(); err != nil {
		return err
	}

	sugar := logger.Sugar()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := circle.NewClient(circleAPIToken)

	result, err := internal.WaitForJobs(
		ctx,
		logger,
		client,
		internal.WaitForJobsOptions{
			ProjectType:     projectType,
			Org:             org,
			Project:         project,
			PipelineNumber:  pipelineNumber,
			WorkflowNames:   commaSeparatedListToSlice(workflow),
			ExcludeJobNames: commaSeparatedListToSlice(exclude),
			JobPrefixes:     commaSeparatedListToSlice(jobPrefix),
			FailOnError:     failOnError,
			WaitDuration:    internal.NewWaitForJobsDuration(waitTime),
		},
	)
	if err != nil {
		return err
	}

	if !result.Failed {
		sugar.Infof("all workflows and jobs finished successfully")
	} else {
		sugar.Errorf("one or more workflows or jobs failed")
		if failOnError {
			fmt.Printf(`

##################################################################################################

%s

`,
				failHeader,
			)

			// report all workflows that have failed
			for _, workflow := range result.FailedWorkflows {
				printWorkflowNameAndURL(workflow.Workflow)
			}

			// report any workflow that has at least one job that has failed
			for _, workflow := range result.PendingWorkflows {
				if len(workflow.FailedJobs) > 0 {
					printWorkflowNameAndURL(workflow.Workflow)
				}
			}

			fmt.Printf(`

%s

##################################################################################################
`,
				failFooter,
			)

			os.Exit(2)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(waitForJobsCmd)

	addWorkflowFlags(waitForJobsCmd)

	waitForJobsCmd.Flags().StringVar(&exclude, "exclude", "", "job or jobs to exclude, comma separated list")
	waitForJobsCmd.Flags().StringVar(&jobPrefix, "job-prefix", "", "job prefix or prefixes to limit filtering to, comma separated list")
	waitForJobsCmd.Flags().BoolVar(&failOnError, "fail-on-error", false, "print human-friendly details about failed workflows and exit with non-zero exit code")
	waitForJobsCmd.Flags().StringVar(&failHeader, "fail-header", "", "additional message header to print before the report of failed CircleCI workflows")
	waitForJobsCmd.Flags().StringVar(&failFooter, "fail-footer", "", "additional message footer to print after the report of failed CircleCI workflows")
	waitForJobsCmd.Flags().DurationVar(&timeout, "timeout", 15*time.Minute, "time out to wait for results")
	waitForJobsCmd.Flags().DurationVar(&waitTime, "wait-time", 10*time.Second, "time out to wait between performing checks (twice as much if >= 3 jobs are still pending)")
}
