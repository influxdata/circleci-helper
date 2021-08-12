package circle

import (
	"context"
	"fmt"
	"net/http"
)

// ClientHTTPError implements error interface for HTTP errors.
type ClientHTTPError struct {
	StatusCode int
}

func newClientHTTPErrorFromResponse(res *http.Response) *ClientHTTPError {
	return &ClientHTTPError{
		StatusCode: res.StatusCode,
	}
}

func (e *ClientHTTPError) Error() string {
	return fmt.Sprintf("invalid HTTP response code: %d", e.StatusCode)
}

// Client provides an interface for calling CircleCI, with the goal of also allowing mocked implementations for tests.
type Client interface {
	// GetPipelineID returns UUID of the pipeline based on project type, org, name and pipeline number.
	GetPipelineID(ctx context.Context, projectType string, org string, project string, pipelineNumber int) (string, error)
	// GetWorkflows retrieves workflows for a specific pipeline ID.
	GetWorkflows(ctx context.Context, pipelineID string) ([]*Workflow, error)
	// GetWorkflowJobs retrieves jobs for a specific workflow ID.
	GetWorkflowJobs(ctx context.Context, workflowID string) ([]*Job, error)
	// GetJobDetails retrieves details for a specific job in a specific project.
	GetJobDetails(ctx context.Context, projectType string, org string, project string, jobNumber int) (*JobDetails, error)
	// GetJobActionOutput retrieves output for a specific action.
	GetJobActionOutput(ctx context.Context, action *JobAction) ([]JobOutputMessage, error)
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
