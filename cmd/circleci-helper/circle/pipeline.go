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

// GetPipelineID returns UUID of the pipeline based on project type, org, name and pipeline number
func GetPipelineID(ctx context.Context, token string, projectType string, org string, project string, pipelineNumber int) (string, error) {
	requestURL := fmt.Sprintf("https://circleci.com/api/v2/project/%s/%s/%s/pipeline/%d", url.PathEscape(projectType), url.PathEscape(org), url.PathEscape(project), pipelineNumber)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return "result", err
	}

	req.SetBasicAuth(token, "")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("invalid HTTP response code: %d", res.StatusCode)
	}

	var response circleGetPipelineIDResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "result", err
	}

	return response.ID, nil
}

// GetWorkflows returns all workflows for specified pipeline id
func GetWorkflows(ctx context.Context, token string, pipelineID string) ([]*Workflow, error) {
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

		req.SetBasicAuth(token, "")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return result, err
		}
		defer res.Body.Close()

		if res.StatusCode >= 400 {
			return result, fmt.Errorf("invalid HTTP response code: %d", res.StatusCode)
		}

		var response circleGetWorkflowsResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return result, err
		}

		// combine results back into result as the API can use pagination
		for _, item := range response.Items {
			result = append(result, item)
		}
		if response.NextPageToken == "" {
			break
		}

		pageToken = response.NextPageToken
	}
	return result, nil
}
