package circle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// helper to deserialize response from CircleCI API
type circleGetPipelineIDResponse struct {
	ID string `json:"id"`
}

// helper to deserialize response from CircleCI API
type circleGetWorkflowsResponse struct {
	Items         []*Workflow `json:"items"`
	NextPageToken string      `json:"next_page_token"`
}

// GetPipelineID returns UUID of the pipeline based on project type, org, name and pipeline number.
func (c *tokenBasedClient) GetPipelineID(ctx context.Context, projectType string, org string, project string, pipelineNumber int) (string, error) {
	requestURL := fmt.Sprintf("https://circleci.com/api/v2/project/%s/%s/%s/pipeline/%d", url.PathEscape(projectType), url.PathEscape(org), url.PathEscape(project), pipelineNumber)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.token, "")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return "", newClientHTTPErrorFromResponse(c.logger, res)
	}

	var response circleGetPipelineIDResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.ID, nil
}

// GetWorkflows retrieves workflows for a specific pipeline ID.
func (c *tokenBasedClient) GetWorkflows(ctx context.Context, pipelineID string) ([]*Workflow, error) {
	var result []*Workflow

	pageToken := ""
	for {
		requestURL := fmt.Sprintf("https://circleci.com/api/v2/pipeline/%s/workflow", pipelineID)
		if pageToken != "" {
			requestURL = fmt.Sprintf("%s?page-token=%s", requestURL, url.QueryEscape(pageToken))
		}

		req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
		if err != nil {
			return result, err
		}

		req.SetBasicAuth(c.token, "")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return result, err
		}
		defer res.Body.Close()

		if res.StatusCode >= 400 {
			return result, newClientHTTPErrorFromResponse(c.logger, res)
		}

		var response circleGetWorkflowsResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return result, err
		}

		// combine results back into result as the API can use pagination
		result = append(result, response.Items...)

		if response.NextPageToken == "" {
			break
		}

		pageToken = response.NextPageToken
	}
	return result, nil
}
