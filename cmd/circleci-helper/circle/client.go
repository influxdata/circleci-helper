package circle

import (
	"context"
)

// Client provides an interface for calling CircleCI, with the goal of also allowing mocked implementations for tests.
type Client interface {
	GetPipelineID(ctx context.Context, projectType string, org string, project string, pipelineNumber int) (string, error)
	GetWorkflows(ctx context.Context, pipelineID string) ([]*Workflow, error)
	GetWorkflowJobs(ctx context.Context, workflowID string) ([]*Job, error)
}

type tokenBasedClient struct {
	token string
}

// NewClient creates a new instance Client that can be used to communicate with CircleCI.
func NewClient(token string) Client {
	return &tokenBasedClient{
		token: token,
	}
}
