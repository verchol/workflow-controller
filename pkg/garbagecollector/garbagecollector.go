package garbagecollector

import (
	"path"
	"time"

	"github.com/golang/glog"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	winformers "github.com/sdminonne/workflow-controller/pkg/client/informers/externalversions"
	wlisters "github.com/sdminonne/workflow-controller/pkg/client/listers/workflow/v1"
	wclientset "github.com/sdminonne/workflow-controller/pkg/client/versioned"

	"github.com/sdminonne/workflow-controller/pkg/controller"
)

const (
	// Interval represent the interval to run Garabge Collection
	Interval time.Duration = 10 * time.Second
)

// GarbageCollector represents a Workflow Garbage Collector.
// It collects orphaned Jobs
type GarbageCollector struct {
	KubeClient     clientset.Interface
	WorkflowClient wclientset.Interface
	WorkflowLister wlisters.WorkflowLister
	WorkflowSynced cache.InformerSynced
}

// NewGarbageCollector builds initializes and returns a GarbageCollector
func NewGarbageCollector(workflowClient wclientset.Interface, kubeClient clientset.Interface, workflowInformerFactory winformers.SharedInformerFactory) *GarbageCollector {
	return &GarbageCollector{
		KubeClient:     kubeClient,
		WorkflowClient: workflowClient,
		WorkflowLister: workflowInformerFactory.Workflow().V1().Workflows().Lister(),
		WorkflowSynced: workflowInformerFactory.Workflow().V1().Workflows().Informer().HasSynced,
	}
}

// CollectWorkflowJobs collect the orphaned jobs. First looking in the workflow informer list
// then retrieve from the API and in case NotFound then remove via DeleteCollection primitive
func (c *GarbageCollector) CollectWorkflowJobs() {
	glog.V(4).Infof("Collecting garbage jobs")
	jobs, err := c.KubeClient.BatchV1().Jobs(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: controller.WorkflowLabelKey,
	})
	if err != nil {
		glog.Errorf("Garbage collector: unable to list workflow jobs to be collected: %v", err)
	}
	collected := make(map[string]struct{})
	for _, job := range jobs.Items {
		workflowName, found := job.Labels[controller.WorkflowLabelKey]
		if !found {
			glog.Errorf("unable to find workflow name for job: %s/%s", job.Namespace, job.Name)
			continue
		}
		if _, done := collected[path.Join(job.Namespace, workflowName)]; done {
			continue // already collected so skip
		}
		if _, err := c.WorkflowLister.Workflows(job.Namespace).Get(workflowName); err == nil || !apierrors.IsNotFound(err) {
			if err != nil {
				glog.Errorf("unable to retrieve workflow %s/%s cache: %v", job.Namespace, workflowName, err)
			}
			continue
		}
		// Workflow couldn't be find in cache. Tyring to get it via APIs.
		if _, err := c.WorkflowClient.Workflow().Workflows(job.Namespace).Get(workflowName, metav1.GetOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				// then remove all the jobs.
				if err := c.KubeClient.Batch().Jobs(job.Namespace).DeleteCollection(controller.CascadeDeleteOptions(0), metav1.ListOptions{
					LabelSelector: controller.WorkflowLabelKey + "=" + workflowName}); err != nil {
					if err != nil {
						glog.Errorf("unable to delete Collection of jobs for workflow %s/%s", job.Namespace, workflowName)
					}
				}
				collected[path.Join(job.Namespace, workflowName)] = struct{}{}
				glog.Infof("Removed all jobs for workflow %s/%s", job.Namespace, workflowName)
				continue
			} else {
				glog.Errorf("unable to retrieve workflow %s/%s for job %s/%s: %v", job.Namespace, workflowName, job.Namespace, job.Name, err)
			}
		}
	}
}
