package framework

import (
	. "github.com/onsi/gomega"

	batch "k8s.io/api/batch/v1"
	batchv2 "k8s.io/api/batch/v2alpha1"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	wapi "github.com/sdminonne/workflow-controller/pkg/api/v1"
)

// BuildAndSetClients builds and initilize workflow and kube client
func BuildAndSetClients() (*rest.RESTClient, *clientset.Clientset) {
	f, err := NewFramework()
	Ω(err).ShouldNot(HaveOccurred())
	Ω(f).ShouldNot(BeNil())

	kubeClient, err := f.kubeClient()
	Ω(err).ShouldNot(HaveOccurred())
	Ω(kubeClient).ShouldNot(BeNil())
	Logf("Check wether Workflow resource is registered...")
	/*
		TODO: check whether CRD is registered
		Eventually(func() bool {
			r, err := client.IsWorkflowRegistered(kubeClient, f.ResourceName, f.ResourceGroup, f.ResourceVersion)
			if err != nil {
				Logf("Error: %v", err)
			}
			return (r && err == nil)
		}, "5s", "1s").Should(BeTrue())
		Logf("It is!")
		resourceName := strings.Join([]string{f.ResourceName, f.ResourceGroup}, ".")
		thirdPartyResource, err := kubeClient.Extensions().ThirdPartyResources().Get(resourceName)
	*/
	workflowClient, err := f.workflowClient()
	Ω(err).ShouldNot(HaveOccurred())
	Ω(workflowClient).ShouldNot(BeNil())
	return workflowClient, kubeClient
}

// NewWorkflow creates a workflow
func NewWorkflow(group, version, name, namespace string, activeDeadlineSeconds *int64) *wapi.Workflow {
	return &wapi.Workflow{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: group + "/" + version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: wapi.WorkflowSpec{
			ActiveDeadlineSeconds: activeDeadlineSeconds,
			Steps: []wapi.WorkflowStep{
				{
					Name: "one",
					JobTemplate: &batchv2.JobTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"workflow": "step_one",
							},
						},
						Spec: batch.JobSpec{
							Template: api.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{
										"foo": "bar",
									},
								},
								Spec: api.PodSpec{
									Containers: []api.Container{
										{
											Name:            "step-one-wait-and-exit",
											Image:           "gcr.io/google_containers/busybox",
											Command:         []string{"sh", "-c", "echo Starting on: $(date); sleep 5; echo Goodbye cruel world at: $(date)"},
											ImagePullPolicy: "IfNotPresent",
										},
									},
									RestartPolicy: "Never",
									DNSPolicy:     "Default",
								},
							},
						},
					},
				},
			},

			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"workflow": "hello",
				},
			},
		},
		//Status: wapi.WorkflowStatus{
		//	Conditions: []wapi.WorkflowCondition{},
		//	Statuses:   []wapi.WorkflowStepStatus{},
		//},
	}
}
