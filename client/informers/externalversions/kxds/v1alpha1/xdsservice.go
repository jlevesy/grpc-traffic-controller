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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/kxds/v1alpha1"
	versioned "github.com/jlevesy/kxds/client/clientset/versioned"
	internalinterfaces "github.com/jlevesy/kxds/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/jlevesy/kxds/client/listers/kxds/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// XDSServiceInformer provides access to a shared informer and lister for
// XDSServices.
type XDSServiceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.XDSServiceLister
}

type xDSServiceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewXDSServiceInformer constructs a new informer for XDSService type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewXDSServiceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredXDSServiceInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredXDSServiceInformer constructs a new informer for XDSService type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredXDSServiceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApiV1alpha1().XDSServices(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApiV1alpha1().XDSServices(namespace).Watch(context.TODO(), options)
			},
		},
		&kxdsv1alpha1.XDSService{},
		resyncPeriod,
		indexers,
	)
}

func (f *xDSServiceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredXDSServiceInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *xDSServiceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kxdsv1alpha1.XDSService{}, f.defaultInformer)
}

func (f *xDSServiceInformer) Lister() v1alpha1.XDSServiceLister {
	return v1alpha1.NewXDSServiceLister(f.Informer().GetIndexer())
}