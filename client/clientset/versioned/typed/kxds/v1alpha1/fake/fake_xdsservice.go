/*
Copyright 2023.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/jlevesy/kxds/api/kxds/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeXDSServices implements XDSServiceInterface
type FakeXDSServices struct {
	Fake *FakeApiV1alpha1
	ns   string
}

var xdsservicesResource = schema.GroupVersionResource{Group: "api.kxds.dev", Version: "v1alpha1", Resource: "xdsservices"}

var xdsservicesKind = schema.GroupVersionKind{Group: "api.kxds.dev", Version: "v1alpha1", Kind: "XDSService"}

// Get takes name of the xDSService, and returns the corresponding xDSService object, and an error if there is any.
func (c *FakeXDSServices) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.XDSService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(xdsservicesResource, c.ns, name), &v1alpha1.XDSService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.XDSService), err
}

// List takes label and field selectors, and returns the list of XDSServices that match those selectors.
func (c *FakeXDSServices) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.XDSServiceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(xdsservicesResource, xdsservicesKind, c.ns, opts), &v1alpha1.XDSServiceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.XDSServiceList{ListMeta: obj.(*v1alpha1.XDSServiceList).ListMeta}
	for _, item := range obj.(*v1alpha1.XDSServiceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested xDSServices.
func (c *FakeXDSServices) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(xdsservicesResource, c.ns, opts))

}

// Create takes the representation of a xDSService and creates it.  Returns the server's representation of the xDSService, and an error, if there is any.
func (c *FakeXDSServices) Create(ctx context.Context, xDSService *v1alpha1.XDSService, opts v1.CreateOptions) (result *v1alpha1.XDSService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(xdsservicesResource, c.ns, xDSService), &v1alpha1.XDSService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.XDSService), err
}

// Update takes the representation of a xDSService and updates it. Returns the server's representation of the xDSService, and an error, if there is any.
func (c *FakeXDSServices) Update(ctx context.Context, xDSService *v1alpha1.XDSService, opts v1.UpdateOptions) (result *v1alpha1.XDSService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(xdsservicesResource, c.ns, xDSService), &v1alpha1.XDSService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.XDSService), err
}

// Delete takes name of the xDSService and deletes it. Returns an error if one occurs.
func (c *FakeXDSServices) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(xdsservicesResource, c.ns, name, opts), &v1alpha1.XDSService{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeXDSServices) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(xdsservicesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.XDSServiceList{})
	return err
}

// Patch applies the patch and returns the patched xDSService.
func (c *FakeXDSServices) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.XDSService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(xdsservicesResource, c.ns, name, pt, data, subresources...), &v1alpha1.XDSService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.XDSService), err
}
