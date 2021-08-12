package circle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// JobDetails describes details for a single job.
type JobDetails struct {
	Steps []JobStep `json:"steps"`
}

// JobStep describes a single job in JobDetails.
type JobStep struct {
	Name    string      `json:"name"`
	Actions []JobAction `json:"actions"`
}

// JobAction describes a single action in a JobStep.
type JobAction struct {
	Name      string `json:"name"`
	Failed    bool   `json:"failed"`
	OutputURL string `json:"output_url"`
	HasOutput bool   `json:"has_output"`
}

type JobOutputMessage struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Time    string `json:"time"`
}

// GetJobDetails retrieves details for a specific job in a specific project.
func (c *tokenBasedClient) GetJobDetails(ctx context.Context, projectType string, org string, project string, jobNumber int) (*JobDetails, error) {
	requestURL := fmt.Sprintf(
		"https://circleci.com/api/v1.1/project/%s/%s/%s/%d",
		url.PathEscape(projectType), url.PathEscape(org), url.PathEscape(project),
		jobNumber,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.token, "")
	req.Header.Add("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, newClientHTTPErrorFromResponse(res)
	}

	var response JobDetails
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetJobActionOutput retrieves output for a specific action.
func (c *tokenBasedClient) GetJobActionOutput(ctx context.Context, action *JobAction) ([]JobOutputMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", action.OutputURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, newClientHTTPErrorFromResponse(res)
	}

	var response []JobOutputMessage

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}
