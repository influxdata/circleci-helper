package cmd

import (
	"context"
	"fmt"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
	"github.com/influxdata/circleci-helper/cmd/circleci-helper/internal"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// workflowErrorsCmd represents the workflow-errors command
var workflowErrorsCmd = &cobra.Command{
	Use:   "workflow-errors",
	Short: "Report all errors for specified workflow",
	Long: `Reports all errors for specified workflow. For example:

circleci-helper workflow-errors --token ... --pipeline ... --workflow "myworkflow" --project-type ...
`,
	Run: func(cmd *cobra.Command, args []string) {
		commandHelper(cmd, args, workflowErrorsMain)
	},
}

func workflowErrorsMain(logger *zap.Logger, cmd *cobra.Command, args []string) error {
	if err := validateWorkflowFlags(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := circle.NewClient(circleAPIToken)

	result, err := internal.WorkflowErrors(ctx, logger, client, internal.WorkflowErrorsOptions{
		ProjectType:    projectType,
		Org:            org,
		Project:        project,
		PipelineNumber: pipelineNumber,
		WorkflowNames:  commaSeparatedListToSlice(workflow),
	})

	if err != nil {
		return err
	}

	for _, failure := range result.Failures {
		fmt.Printf("Failed to run workflow %s job %s at step %s action %s:\n%s\n\n----\n", failure.Workflow.Name, failure.Job.Name, failure.StepName, failure.ActionName, failure.Messages)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(workflowErrorsCmd)

	addWorkflowFlags(workflowErrorsCmd)
}
