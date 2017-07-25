/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"time"

	"k8s.io/kubernetes/pkg/api"

	"k8s.io/kubernetes/pkg/api/unversioned"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"

	wapi "github.com/sdminonne/workflow-controller/pkg/api"
)

// WorkflowsNamespacer has methods to work with Workflow resources in a namespace
type WorkflowsNamespacer interface {
	Workflows(namespace string) WorkflowInterface
}

// Interface is just an alias for WorkflowNamespacer
type Interface WorkflowsNamespacer

// WorkflowInterface exposes methods to work on Workflow resources.
type WorkflowInterface interface {
	Create(*wapi.Workflow) (*wapi.Workflow, error)
	List(options api.ListOptions) (*wapi.WorkflowList, error)
	Get(name string) (*wapi.Workflow, error)
	Update(workflow *wapi.Workflow) (*wapi.Workflow, error)
	Delete(name string, options *api.DeleteOptions) error

	//Watch(options api.ListOptions) (watch.Interface, error)
}

// Client implements a workflow client
type Client struct {
	*dynamic.Client
	restResource     string
	groupVersionKind *unversioned.GroupVersionKind
}

// Workflows returns a Workflows
func (c Client) Workflows(ns string) WorkflowInterface {
	return newWorkflows(c, ns)
}

// CreateWorkflowResource creates the Workflow resource as a Kubernetes CR
func CreateWorkflowResource() *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "workflows.amadeus.net"},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "amadeus.net",
			Version: "v1beta1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     "workflows",
				Singular:   "workflow",
				Kind:       "Workflow",
				ShortNames: []string{"wfl"},
				ListKind:   "WorkflowsList",
			},
			Scope: apiextensionsv1beta1.ClusterScoped,
		},
	}
}

// CreateWorkflowClient creates the client to handle Workflows
func CreateWorkflowClient(crd *apiextensionsv1beta1.CustomResourceDefinition, apiExtensionsClient clientset.Interface, clientPool dynamic.ClientPool) (*dynamic.Client, error) {
	_, err := apiExtensionsClient.Apiextensions().CustomResourceDefinitions().Create(crd)
	if err != nil {
		return nil, err
	}

	// wait until the resource appears in discovery
	err = wait.PollImmediate(500*time.Millisecond, 30*time.Second, func() (bool, error) {
		resourceList, err := apiExtensionsClient.Discovery().ServerResourcesForGroupVersion(crd.Spec.Group + "/" + crd.Spec.Version)
		if err != nil {
			return false, nil
		}
		for _, resource := range resourceList.APIResources {
			if resource.Name == crd.Spec.Names.Plural {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return clientPool.ClientForGroupVersionResource(schema.GroupVersionResource{Group: crd.Spec.Group, Version: crd.Spec.Version, Resource: crd.Spec.Names.Plural})

}

// workflows implements WorkflowsNamespacer interface
type workflows struct {
	c        Client
	ns       string
	gvk      *unversioned.GroupVersionKind
	resource *unversioned.APIResource
}

// newWorkflows returns a workflows
func newWorkflows(c Client, ns string) *workflows {
	return &workflows{
		c:        c,
		ns:       ns,
		gvk:      c.groupVersionKind,
		resource: &unversioned.APIResource{Name: c.restResource, Namespaced: len(ns) != 0},
	}
}

// Ensure statically that workflows implements WorkflowInterface.
var _ WorkflowInterface = &workflows{}

// Create creates a Workflow
func (w *workflows) Create(workflow *wapi.Workflow) (*wapi.Workflow, error) {
	return nil, nil
}

// List returns a list of workflows that match the label and field selectors.
func (w *workflows) List(options api.ListOptions) (*wapi.WorkflowList, error) {
	return &wapi.WorkflowList{}, nil
}

// Get returns information about a particular workflow.
func (w *workflows) Get(name string) (*wapi.Workflow, error) {
	workflow := &wapi.Workflow{}
	return &wapi.Workflow{}, nil
}

// Update updates an existing workflow. TODO: implement via PATCH
func (w *workflows) Update(workflow *wapi.Workflow) (*wapi.Workflow, error) {
	updatedWorkflow := &wapi.Workflow{}
	return updatedWorkflow, nil

}

// Delete deletes a workflow, returns error if one occurs.
func (w *workflows) Delete(name string, options *api.DeleteOptions) error {
	return nil
}

// IsWorkflowRegistered returns wheter or not the Workflow with specific group, name and version is registered
func IsWorkflowRegistered(crd *apiextensionsv1beta1.CustomResourceDefinition, apiExtensionsClient clientset.Interface, clientPool dynamic.ClientPool) (bool, error) {
	return true, nil
}

/*
// Watch returns a watch.Interface that watches the requested workflows.
func (w *workflows) Watch(options api.ListOptions) (watch.Interface, error) {
	glog.V(6).Infof("Watching workflows...")
	watcher := NewWatcher()
	optionsV1 := v1.ListOptions{}
	v1.Convert_api_ListOptions_To_v1_ListOptions(&options, &optionsV1, nil)
	go wait.Until(func() {
		unstructuredWatcher, err := w.c.Resource(w.resource, w.ns).Watch(&optionsV1)
		if err != nil {
			glog.Errorf("unable to watch workflow: %v", err)
			return
		}
		event, ok := <-unstructuredWatcher.ResultChan()
		if !ok {
			glog.Errorf("Watching workflows: channel closed")
			return
		}

		glog.V(6).Infof("Got event... %s", event.Type)
		if event.Type == watch.Error {
			glog.Errorf("watcher error: %v", apierrs.FromObject(event.Object))
			return
		}
		u, ok := event.Object.(*runtime.Unstructured)
		if !ok {
			glog.Errorf("unable to cast watched object to runtime.Unstructured")
			return
		}

		workflow, err := wcodec.UnstructuredToWorkflow(u)
		if err != nil {
			glog.Errorf("Unable to decode runtime.Unstructured object %v", u)
			return
		}
		optionsV1.ResourceVersion = workflow.ResourceVersion

		watcher.Result <- watch.Event{
			Type:   event.Type,
			Object: workflow,
		}
		glog.V(6).Infof("Queued event in Workflow watcher...: %s", event.Type)
	}, 25*time.Millisecond, wait.NeverStop)
	return watcher, nil
}
*/
