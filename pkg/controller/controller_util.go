package controller

import (
	"fmt"

	batch "k8s.io/api/batch/v1"
	batchv2 "k8s.io/api/batch/v2alpha1"
	api "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/labels"

	wapi "github.com/sdminonne/workflow-controller/pkg/api/v1"
)

// IsWorkflowFinished checks wether a workflow is finished. A workflow is finished if one of its condition is Complete or Failed.
func IsWorkflowFinished(w *wapi.Workflow) bool {
	for _, c := range w.Status.Conditions {
		if c.Status == api.ConditionTrue && (c.Type == wapi.WorkflowComplete || c.Type == wapi.WorkflowFailed) {
			return true
		}
	}
	return false
}

// RemoveStepFromSpec remove Step from Workflow Spec
func RemoveStepFromSpec(w *wapi.Workflow, stepName string) error {
	for i := range w.Spec.Steps {
		if w.Spec.Steps[i].Name == stepName {
			w.Spec.Steps = w.Spec.Steps[:i+copy(w.Spec.Steps[i:], w.Spec.Steps[i+1:])]
			return nil
		}
	}
	return fmt.Errorf("unable to find step %q in workflow", stepName)
}

// GetStepByName returns a pointer to Workflow Step
func GetStepByName(w *wapi.Workflow, stepName string) *wapi.WorkflowStep {
	for i := range w.Spec.Steps {
		if w.Spec.Steps[i].Name == stepName {
			return &w.Spec.Steps[i]
		}
	}
	return nil
}

// GetStepStatusByName returns a pointer to Workflow Step
func GetStepStatusByName(w *wapi.Workflow, stepName string) *wapi.WorkflowStepStatus {
	for i := range w.Status.Statuses {
		if w.Status.Statuses[i].Name == stepName {
			return &w.Status.Statuses[i]
		}
	}
	return nil
}

// createWorkflowJobLabelSelector creates label selector to select the jobs related to a workflow, stepName
func createWorkflowJobLabelSelector(workflow *wapi.Workflow, template *batchv2.JobTemplateSpec, stepName string) labels.Selector {
	set, err := getJobLabelsSet(workflow, template, stepName)
	if err != nil {
		return nil
	}
	return labels.SelectorFromSet(set)
}

// IsJobFinished checks whether a Job is finished
func IsJobFinished(j *batch.Job) bool {
	for _, c := range j.Status.Conditions {
		if (c.Type == batch.JobComplete || c.Type == batch.JobFailed) && c.Status == api.ConditionTrue {
			return true
		}
	}
	return false
}
