package circle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Workflow describes a single CircleCI workflow.
// These can be extended to map  more fields from responses as needed.
type Workflow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// helper to deserialize response from CircleCI API
type circleGetWorkflowJobsResponse struct {
	Items         []*Job `json:"items"`
	NextPageToken string `json:"next_page_token"`
}

// GetWorkflowJobs retrieves jobs for a specific workflow ID.
func (c *tokenBasedClient) GetWorkflowJobs(ctx context.Context, workflowID string) ([]*Job, error) {
	var result []*Job

	pageToken := ""
	for {
		requestURL := fmt.Sprintf("https://circleci.com/api/v2/workflow/%s/job", url.PathEscape(workflowID))
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

		var response circleGetWorkflowJobsResponse
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

// WorkflowFinished returns whether specified workflow has finished and is no longer in progress.
func WorkflowFinished(workflow *Workflow) bool {
	return WorkflowFailed(workflow) || workflow.Status == "success"
}

// WorkflowFailed returns whether specified workflow has failed.
func WorkflowFailed(workflow *Workflow) bool {
	return workflow.Status == "failed" ||
		workflow.Status == "error" ||
		workflow.Status == "canceled" ||
		workflow.Status == "unauthorized"
}
