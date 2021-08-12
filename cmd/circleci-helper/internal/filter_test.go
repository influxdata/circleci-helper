package internal

import (
	"testing"

	"github.com/influxdata/circleci-helper/cmd/circleci-helper/circle"
)

func Test_uniqueWorkflows(t *testing.T) {
	m := newMockCircleClientWithData("success", "success", "success", "success")
	resultWorkflows := uniqueWorkflows(m.workflowsMap["456"])
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
}

func Test_filterWorkflowWrapper(t *testing.T) {
	workflow := &circle.Workflow{
		Name: "Test",
	}
	for _, test := range []struct {
		name      string
		keepNames []string
		expect    bool
	}{
		{
			name:   "empty keepNames",
			expect: true,
		},
		{
			name:      "invalid name",
			keepNames: []string{"Wrong"},
			expect:    false,
		},
		{
			name:      "valid name",
			keepNames: []string{"Test"},
			expect:    true,
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			if want, got := test.expect, filterWorkflowWrapper(test.keepNames)(workflow); want != got {
				tt.Errorf("invalid result; want %v, got %v", want, got)
			}
		})
	}
}

func Test_filterJobWrapper(t *testing.T) {
	workflow := &circle.Job{
		Name: "Prefix-Test",
	}
	for _, test := range []struct {
		name            string
		excludeJobNames []string
		jobPrefixes     []string
		expect          bool
	}{
		{
			name:   "no settings",
			expect: true,
		},
		{
			name:        "invalid prefix",
			jobPrefixes: []string{"Invalid-"},
			expect:      false,
		},
		{
			name:        "valid prefix",
			jobPrefixes: []string{"Prefix-"},
			expect:      true,
		},
		{
			name:            "not excluded",
			excludeJobNames: []string{"Invalid"},
			expect:          true,
		},
		{
			name:            "excluded",
			excludeJobNames: []string{"Prefix-Test"},
			expect:          false,
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			if want, got := test.expect, filterJobWrapper(test.excludeJobNames, test.jobPrefixes)(workflow); want != got {
				tt.Errorf("invalid result; want %v, got %v", want, got)
			}
		})
	}
}
