package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/internal"
)

var pipelineNumber int
var projectType string
var org string
var project string
var workflow string
var exclude string
var failOnError bool

// waitForJobsCmd represents the waitForJobs command
var waitForJobsCmd = &cobra.Command{
	Use:   "wait-for-jobs",
	Short: "Wait for other job or jobs in specified workflow",
	Long: `Wait for one or more other jobs in specified workflow. For example:

circleci-helper wait-for-jobs --token ... --pipeline ... --workflow "myworkflow" --project-type ... --exclude "my-finalize-job"
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateFlags(); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing options: %v\n", err)
			os.Exit(1)
		}

		success, err := internal.WaitForJobs(
			circleAPIToken,
			projectType, org, project,
			pipelineNumber,
			commaSeparatedListToArray(workflow),
			commaSeparatedListToArray(exclude),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
			os.Exit(1)
		}

		if success {
			fmt.Printf("OK   All workflows and jobs finished successfully\n")
		} else {
			fmt.Printf("FAIL One or more workflows or jobs failed\n")
			if failOnError {
				os.Exit(2)
			}
		}
	},
}

func validateFlags() error {
	if org == "" {
		return fmt.Errorf("org must be specified")
	}
	if project == "" {
		return fmt.Errorf("project must be specified")
	}
	if pipelineNumber == 0 {
		return fmt.Errorf("pipeline-number must be specified")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(waitForJobsCmd)

	waitForJobsCmd.Flags().IntVar(&pipelineNumber, "pipeline-number", 0, "pipeline number")
	waitForJobsCmd.Flags().StringVar(&projectType, "project-type", "github", "project type (i.e. github)")
	waitForJobsCmd.Flags().StringVar(&org, "org", "", "organization")
	waitForJobsCmd.Flags().StringVar(&project, "project", "", "project")
	waitForJobsCmd.Flags().StringVar(&workflow, "workflow", "", "workflow names to limit to, comma separated list")
	waitForJobsCmd.Flags().StringVar(&exclude, "exclude", "", "job or jobs to exclude, comma separated list")
	waitForJobsCmd.Flags().BoolVar(&failOnError, "fail-on-error", false, "return non-zero exit code if one or more jobs have failed")
}
