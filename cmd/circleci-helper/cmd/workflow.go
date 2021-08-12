package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// common flags and variables for workflow-related commands
var pipelineNumber int
var projectType string
var org string
var project string
var workflow string

// addWorkflowFlags adds common workflow-related flags to the given command.
func addWorkflowFlags(command *cobra.Command) {
	command.Flags().IntVar(&pipelineNumber, "pipeline-number", 0, "pipeline number")
	command.Flags().StringVar(&projectType, "project-type", "github", "project type (i.e. github)")
	command.Flags().StringVar(&org, "org", "", "organization")
	command.Flags().StringVar(&project, "project", "", "project")
	command.Flags().StringVar(&workflow, "workflow", "", "workflow names to limit to, comma separated list")
}

// validateWorkflowFlags validates flags common for workflow-related commands.
func validateWorkflowFlags() error {
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
