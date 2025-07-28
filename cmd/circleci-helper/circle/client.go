package circle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// ClientHTTPError implements error interface for HTTP errors.
type ClientHTTPError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message,omitempty"`
}

func newClientHTTPErrorFromResponse(logger *zap.Logger, res *http.Response) *ClientHTTPError {
	var ce ClientHTTPError
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&ce); err != nil {
		logger.Warn("failed to decode CircleCI API error response", zap.Error(err))
	}
	ce.StatusCode = res.StatusCode
	return &ce
}

func (e *ClientHTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
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
	logger *zap.Logger
	token  string
}

// NewClient creates a new instance Client that can be used to communicate with CircleCI.
func NewClient(logger *zap.Logger, token string) Client {
	return &tokenBasedClient{
		logger: logger,
		token:  token,
	}
}
