package internal

import (
	"context"
	"fmt"
	"testing"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
)

type mockCircleClient struct {
	projectType   string
	org           string
	project       string
	pipelineIDMap map[int]string
	workflowsMap  map[string][]*circle.Workflow
	jobsMap       map[string][]*circle.Job
	jobDetailsMap map[int]*circle.JobDetails
	jobOutputMap  map[string][]circle.JobOutputMessage
}

func newMockCircleClient(projectType, org, project string) *mockCircleClient {
	return &mockCircleClient{
		projectType:   projectType,
		org:           org,
		project:       project,
		pipelineIDMap: map[int]string{},
		workflowsMap:  map[string][]*circle.Workflow{},
		jobsMap:       map[string][]*circle.Job{},
		jobDetailsMap: map[int]*circle.JobDetails{},
		jobOutputMap:  map[string][]circle.JobOutputMessage{},
	}
}

func (m *mockCircleClient) addPipeline(number int, id string) {
	m.pipelineIDMap[number] = id
}

func (m *mockCircleClient) addWorkflows(pipelineID string, workflows []*circle.Workflow) {
	m.workflowsMap[pipelineID] = workflows
}

func (m *mockCircleClient) addJobs(workflowID string, jobs []*circle.Job) {
	m.jobsMap[workflowID] = jobs
}

func (m *mockCircleClient) addJobDetails(jobID int, details *circle.JobDetails) {
	m.jobDetailsMap[jobID] = details
}

func (m *mockCircleClient) addJobOutput(job *circle.JobAction, output []circle.JobOutputMessage) {
	m.jobOutputMap[job.OutputURL] = output
}

func (m *mockCircleClient) addJobOutputString(job *circle.JobAction, message string) {
	m.jobOutputMap[job.OutputURL] = []circle.JobOutputMessage{
		{
			Message: message,
			Time:    "2021-01-01T00:00:00.000Z",
			Type:    "out",
		},
	}
}

func (m *mockCircleClient) GetPipelineID(ctx context.Context, projectType string, org string, project string, pipelineNumber int) (string, error) {
	if m.projectType != projectType || m.org != org || m.project != project {
		return "", fmt.Errorf("invalid project info")
	}
	res, ok := m.pipelineIDMap[pipelineNumber]
	if !ok {
		return "", fmt.Errorf("invalid pipeline number")
	}
	return res, nil
}

func (m *mockCircleClient) GetWorkflows(ctx context.Context, pipelineID string) ([]*circle.Workflow, error) {
	res, ok := m.workflowsMap[pipelineID]
	if !ok {
		return nil, fmt.Errorf("invalid pipelineID")
	}
	return res, nil
}

func (m *mockCircleClient) GetWorkflowJobs(ctx context.Context, workflowID string) ([]*circle.Job, error) {
	res, ok := m.jobsMap[workflowID]
	if !ok {
		return nil, fmt.Errorf("invalid workflowID")
	}
	return res, nil
}

func (m *mockCircleClient) GetJobDetails(ctx context.Context, projectType string, org string, project string, jobNumber int) (*circle.JobDetails, error) {
	if m.projectType != projectType || m.org != org || m.project != project {
		return nil, fmt.Errorf("invalid project info")
	}
	res, ok := m.jobDetailsMap[jobNumber]
	if !ok {
		return nil, fmt.Errorf("invalid jobNumber")
	}
	return res, nil
}

func (m *mockCircleClient) GetJobActionOutput(ctx context.Context, action *circle.JobAction) ([]circle.JobOutputMessage, error) {
	res, ok := m.jobOutputMap[action.OutputURL]
	if !ok {
		return nil, fmt.Errorf("invalid action's OutputURL")
	}
	return res, nil
}

func newMockCircleClientWithData(workflow1Status, workflow2Status, job1Status, job2Status string) *mockCircleClient {
	m := newMockCircleClient("github", "influxdata", "testproject")
	m.addPipeline(123, "456")
	m.addPipeline(987, "654")

	m.addWorkflows("456", []*circle.Workflow{
		{
			ID:        "456-1",
			Name:      "test-workflow-1",
			Status:    workflow1Status,
			CreatedAt: "2021-01-01T00:00:00.000Z",
		},
		{
			ID:        "456-2",
			Name:      "test-workflow-1",
			Status:    workflow1Status,
			CreatedAt: "2021-01-02T00:00:00.000Z",
		},
		{
			ID:        "456-3",
			Name:      "test-workflow-1",
			Status:    workflow1Status,
			CreatedAt: "2021-01-01T12:00:00.000Z",
		},
		{
			ID:        "456-4",
			Name:      "test-workflow-2",
			Status:    workflow2Status,
			CreatedAt: "2021-01-04T00:00:00.000Z",
		},
	})
	m.addJobs("456-1", []*circle.Job{
		{
			ID:     "456-1-1",
			Name:   "test-job-1-1",
			Status: job1Status,
		},
		{
			ID:     "456-1-2",
			Name:   "test-job-1-2",
			Status: job1Status,
		},
	})
	m.addJobs("456-2", []*circle.Job{
		{
			ID:     "456-2-1",
			Name:   "test-job-2-1",
			Status: job1Status,
		},
		{
			ID:     "456-2-2",
			Name:   "test-job-2-1",
			Status: job1Status,
		},
	})
	m.addJobs("456-3", []*circle.Job{
		{
			ID:     "456-3-1",
			Name:   "test-job-3-1",
			Status: job1Status,
		},
	})
	m.addJobs("456-4", []*circle.Job{
		{
			ID:     "456-4-1",
			Name:   "test-job-4-1",
			Status: job2Status,
		},
	})

	m.addWorkflows("654", []*circle.Workflow{})

	return m
}

func Test_getLatestWorkflows(t *testing.T) {
	m := newMockCircleClientWithData("success", "success", "success", "success")
	ctx := context.Background()

	resultWorkflows, err := getLatestWorkflows(
		ctx, m, "456",
		func(workflow *circle.Workflow) bool { return true },
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want, got := 2, len(resultWorkflows); want != got {
		t.Errorf("invalid number of workflows returned; want %v, got %v", want, got)
	}

	// first result should be the one with the latest createdAt, hence 456-4
	if want, got := resultWorkflows[0].ID, "456-4"; want != got {
		t.Errorf("invalid workflow; want %v, got %v", want, got)
	}

	// second result should be latest test-workflow-1, hence 456-2
	if want, got := resultWorkflows[1].ID, "456-2"; want != got {
		t.Errorf("invalid workflow; want %v, got %v", want, got)
	}

	// validate filter method
	resultWorkflows, err = getLatestWorkflows(
		ctx, m, "456",
		func(workflow *circle.Workflow) bool { return workflow.Name == "test-workflow-1" },
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want, got := 1, len(resultWorkflows); want != got {
		t.Errorf("invalid number of workflows returned; want %v, got %v", want, got)
	}

	// only result should be latest test-workflow-1, hence 456-2
	if want, got := resultWorkflows[0].ID, "456-2"; want != got {
		t.Errorf("invalid workflow; want %v, got %v", want, got)
	}
}

func Test_checkWorkflowsStatus(t *testing.T) {
	for _, test := range []struct {
		name                       string
		workflowStatus             string
		jobStatus                  string
		succeededWorkflowJobsCount []int
		pendingWorkflowJobsCount   []int
		failedWorkflowJobsCount    []int
		succeededJobDetails        bool
		failedJobDetails           bool
		pendingJobDetails          bool
		expectedFinished           bool
		expectedFailed             bool
	}{
		{
			name:           "validate successful flows without retrieving job details",
			workflowStatus: "success", jobStatus: "success",
			succeededWorkflowJobsCount: []int{0, 0},

			expectedFinished: true,
		},
		{
			name:           "validate successful flows without retrieving job details",
			workflowStatus: "success", jobStatus: "success",
			succeededJobDetails:        true,
			succeededWorkflowJobsCount: []int{1, 2},

			expectedFinished: true,
		},

		{
			name:           "validate failed flows without retrieving job details",
			workflowStatus: "failed", jobStatus: "failed",
			failedWorkflowJobsCount: []int{0, 0},

			expectedFinished: true,
			expectedFailed:   true,
		},
		{
			name:           "validate failed flows without retrieving job details",
			workflowStatus: "failed", jobStatus: "failed",
			failedJobDetails:        true,
			failedWorkflowJobsCount: []int{1, 2},

			expectedFinished: true,
			expectedFailed:   true,
		},

		{
			name:           "validate pending flows without retrieving job details",
			workflowStatus: "running", jobStatus: "running",
			pendingWorkflowJobsCount: []int{0, 0},
		},
		{
			name:           "validate pending flows without retrieving job details",
			workflowStatus: "running", jobStatus: "running",
			pendingJobDetails:        true,
			pendingWorkflowJobsCount: []int{1, 2},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			// helper to validate number of workflows returned for specified state and jobs count, using the
			// getJobCount callback to get the job count as the field for jobs differs depending on type of check
			compareWorkflowAndJobs := func(name string, details []*WorkflowDetails, workflowJobsCount []int, getJobCount func(details *WorkflowDetails) int) {
				if want, got := len(workflowJobsCount), len(details); want != got {
					tt.Errorf("invalid number of %s workflows returned; want %v, got %v", name, want, got)
					return
				}

				// validate the expected number of jobs for each workflow
				for i, expectedCount := range workflowJobsCount {
					if want, got := expectedCount, getJobCount(details[i]); want != got {
						tt.Errorf("invalid number of %s jobs in %d returned; want %v, got %v", name, i, want, got)
					}
				}
			}

			m := newMockCircleClientWithData(test.workflowStatus, test.workflowStatus, test.jobStatus, test.jobStatus)
			ctx := context.Background()
			// verify checkWorkflowsStatus does not retrieve job details if not requested
			resultWorkflows, err := checkWorkflowsStatus(
				ctx, m, "456",
				checkWorkflowStatusOpts{
					succeededJobDetails: test.succeededJobDetails,
					failedJobDetails:    test.failedJobDetails,
					pendingJobDetails:   test.pendingJobDetails,
				},
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// validate the Finished field
			if want, got := test.expectedFinished, resultWorkflows.Finished; want != got {
				tt.Errorf("invalid value for Finished; want %v, got %v", want, got)
			}

			// validate the Failed field
			if want, got := test.expectedFailed, resultWorkflows.Failed; want != got {
				tt.Errorf("invalid value for Finished; want %v, got %v", want, got)
			}

			// validate each of workflows and jobs
			compareWorkflowAndJobs("succeeded", resultWorkflows.SucceededWorkflows, test.succeededWorkflowJobsCount, func(details *WorkflowDetails) int { return len(details.SucceededJobs) })
			compareWorkflowAndJobs("pending", resultWorkflows.PendingWorkflows, test.pendingWorkflowJobsCount, func(details *WorkflowDetails) int { return len(details.PendingJobs) })
			compareWorkflowAndJobs("failed", resultWorkflows.FailedWorkflows, test.failedWorkflowJobsCount, func(details *WorkflowDetails) int { return len(details.FailedJobs) })
		})
	}
}
