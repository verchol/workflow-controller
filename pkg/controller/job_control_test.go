package controller

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	api "k8s.io/api/core/v1"
)

/*
func TestWorkfowJobControl(t *testing.T) {
	expectedJob := &batchv1.Job{}
	fakeClient := fake.NewSimpleClientset(expectedJob)
	fakeRecorder := &record.FakeRecorder{}
	jc := WorkflowJobControl{fakeClient, fakeRecorder}

	jobTemplateSpec := &batchv2.JobTemplateSpec{}
	job, err := jc.CreateJob("ns", jobTemplateSpec, nil, "step")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(job, expectedJob) {
		t.Errorf("Unexpected job")
	}
}
*/

func TestIsJobFinished(t *testing.T) {
	testcases := map[string]struct {
		Job      *batchv1.Job
		finished bool
	}{
		"job finished": {
			Job: &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobComplete,
							Status: api.ConditionTrue,
						},
					},
				},
			},
			finished: true,
		},
		"job not finished": {
			Job: &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{},
				},
			},
			finished: false,
		},
	}
	for name, test := range testcases {
		if IsJobFinished(test.Job) != test.finished {
			t.Errorf("%s bad value", name)
		}
	}

}
