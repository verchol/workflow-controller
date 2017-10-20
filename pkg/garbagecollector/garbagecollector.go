package garbagecollector

import (
	"github.com/golang/glog"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"

	wclientset "github.com/sdminonne/workflow-controller/pkg/client/versioned"
	"github.com/sdminonne/workflow-controller/pkg/controller"
)

// GarbageCollector represents a Workflow Garbage Collector.
// It collects orphaned Jobs
type GarbageCollector struct {
	KubeClient     clientset.Interface
	WorkflowClient wclientset.Interface
}

// NewGarbageCollector builds initializes and returns a GarbageCollector
func NewGarbageCollector(workflowClient wclientset.Interface, kubeClient clientset.Interface) *GarbageCollector {
	return &GarbageCollector{
		KubeClient:     kubeClient,
		WorkflowClient: workflowClient,
	}
}

// CollectWorkflowJobs collect the orphaned jobs
func (c *GarbageCollector) CollectWorkflowJobs() {
	glog.V(6).Infof("Collecting garbage jobs")
	jobs, err := c.KubeClient.BatchV1().Jobs(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: controller.WorkflowLabelKey,
	})
	if err != nil {
		glog.Errorf("Garbage collector: unable to list workflow jobs to be collected: %v", err)
	}
	for _, job := range jobs.Items {
		workflowName, found := job.Labels[controller.WorkflowLabelKey]
		if !found {
			glog.Errorf("Unable to find workflow name for job: %s/%s", job.Namespace, job.Name)
			continue
		}
		glog.Infof("Found job %s/%s", job.Namespace, job.Name)
		if _, err := c.WorkflowClient.Workflow().Workflows(job.Namespace).Get(workflowName, metav1.GetOptions{}); err != nil {
			if !apierrors.IsNotFound(err) {
				glog.Errorf("unable to retrieve workflow %s/%s for job %s/%s: %v", job.Namespace, workflowName, job.Namespace, job.Name, err)
				continue
			}
			if err := c.KubeClient.Batch().Jobs(job.Namespace).Delete(job.Name, &metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
				glog.Errorf("GC: Unable to delete %s/%s: %v ", job.Namespace, job.Name, err)
				continue
			}
			glog.Infof("Job %s/%s collected from Garbage Collector", job.Namespace, job.Name)
		}
	}
}
