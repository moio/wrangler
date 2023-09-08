/*
Copyright The Kubernetes Authors.

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

// Code generated by main. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/lasso/pkg/grapher"
	"github.com/rancher/wrangler/pkg/generic"
	v1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type LeaseHandler func(string, *v1.Lease) (*v1.Lease, error)

type LeaseController interface {
	generic.ControllerMeta
	LeaseClient

	OnChange(ctx context.Context, name string, sync LeaseHandler)
	OnRemove(ctx context.Context, name string, sync LeaseHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() LeaseCache
}

type LeaseClient interface {
	Create(*v1.Lease) (*v1.Lease, error)
	Update(*v1.Lease) (*v1.Lease, error)

	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.Lease, error)
	List(namespace string, opts metav1.ListOptions) (*v1.LeaseList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Lease, err error)
}

type LeaseCache interface {
	Get(namespace, name string) (*v1.Lease, error)
	List(namespace string, selector labels.Selector) ([]*v1.Lease, error)

	AddIndexer(indexName string, indexer LeaseIndexer)
	GetByIndex(indexName, key string) ([]*v1.Lease, error)
}

type LeaseIndexer func(obj *v1.Lease) ([]string, error)

type leaseController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewLeaseController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) LeaseController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &leaseController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromLeaseHandlerToHandler(sync LeaseHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.Lease
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.Lease))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *leaseController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.Lease))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateLeaseDeepCopyOnChange(client LeaseClient, obj *v1.Lease, handler func(obj *v1.Lease) (*v1.Lease, error)) (*v1.Lease, error) {
	if obj == nil {
		return obj, nil
	}

	copyObj := obj.DeepCopy()
	newObj, err := handler(copyObj)
	if newObj != nil {
		copyObj = newObj
	}
	if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
		return client.Update(copyObj)
	}

	return copyObj, err
}

func (c *leaseController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *leaseController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(handler)})
		return generic.NewRemoveHandler(name, c.Updater(), handler)(key, obj)
	})
}

func (c *leaseController) OnChange(ctx context.Context, name string, sync LeaseHandler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(sync)})
		return FromLeaseHandlerToHandler(sync)(key, obj)
	})
}

func (c *leaseController) OnRemove(ctx context.Context, name string, sync LeaseHandler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(sync)})
		return generic.NewRemoveHandler(name, c.Updater(), FromLeaseHandlerToHandler(sync))(key, obj)
	})
}

func (c *leaseController) Enqueue(namespace, name string) {
	c.controller.Enqueue(namespace, name)
}

func (c *leaseController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controller.EnqueueAfter(namespace, name, duration)
}

func (c *leaseController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *leaseController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *leaseController) Cache() LeaseCache {
	return &leaseCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *leaseController) Create(obj *v1.Lease) (*v1.Lease, error) {
	result := &v1.Lease{}
	return result, c.client.Create(context.TODO(), obj.Namespace, obj, result, metav1.CreateOptions{})
}

func (c *leaseController) Update(obj *v1.Lease) (*v1.Lease, error) {
	result := &v1.Lease{}
	return result, c.client.Update(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *leaseController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), namespace, name, *options)
}

func (c *leaseController) Get(namespace, name string, options metav1.GetOptions) (*v1.Lease, error) {
	result := &v1.Lease{}
	return result, c.client.Get(context.TODO(), namespace, name, result, options)
}

func (c *leaseController) List(namespace string, opts metav1.ListOptions) (*v1.LeaseList, error) {
	result := &v1.LeaseList{}
	return result, c.client.List(context.TODO(), namespace, result, opts)
}

func (c *leaseController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), namespace, opts)
}

func (c *leaseController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*v1.Lease, error) {
	result := &v1.Lease{}
	return result, c.client.Patch(context.TODO(), namespace, name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type leaseCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *leaseCache) Get(namespace, name string) (*v1.Lease, error) {
	obj, exists, err := c.indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v1.Lease), nil
}

func (c *leaseCache) List(namespace string, selector labels.Selector) (ret []*v1.Lease, err error) {

	err = cache.ListAllByNamespace(c.indexer, namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Lease))
	})

	return ret, err
}

func (c *leaseCache) AddIndexer(indexName string, indexer LeaseIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.Lease))
		},
	}))
}

func (c *leaseCache) GetByIndex(indexName, key string) (result []*v1.Lease, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1.Lease, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1.Lease))
	}
	return result, nil
}
