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
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/kv"
	v1 "k8s.io/api/core/v1"
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

type ServiceHandler func(string, *v1.Service) (*v1.Service, error)

type ServiceController interface {
	generic.ControllerMeta
	ServiceClient

	OnChange(ctx context.Context, name string, sync ServiceHandler)
	OnRemove(ctx context.Context, name string, sync ServiceHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() ServiceCache
}

type ServiceClient interface {
	Create(*v1.Service) (*v1.Service, error)
	Update(*v1.Service) (*v1.Service, error)
	UpdateStatus(*v1.Service) (*v1.Service, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.Service, error)
	List(namespace string, opts metav1.ListOptions) (*v1.ServiceList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Service, err error)
}

type ServiceCache interface {
	Get(namespace, name string) (*v1.Service, error)
	List(namespace string, selector labels.Selector) ([]*v1.Service, error)

	AddIndexer(indexName string, indexer ServiceIndexer)
	GetByIndex(indexName, key string) ([]*v1.Service, error)
}

type ServiceIndexer func(obj *v1.Service) ([]string, error)

type serviceController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewServiceController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) ServiceController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &serviceController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromServiceHandlerToHandler(sync ServiceHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.Service
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.Service))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *serviceController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.Service))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateServiceDeepCopyOnChange(client ServiceClient, obj *v1.Service, handler func(obj *v1.Service) (*v1.Service, error)) (*v1.Service, error) {
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

func (c *serviceController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *serviceController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(handler)})
		return generic.NewRemoveHandler(name, c.Updater(), handler)(key, obj)
	})
}

func (c *serviceController) OnChange(ctx context.Context, name string, sync ServiceHandler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(sync)})
		return FromServiceHandlerToHandler(sync)(key, obj)
	})
}

func (c *serviceController) OnRemove(ctx context.Context, name string, sync ServiceHandler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(sync)})
		return generic.NewRemoveHandler(name, c.Updater(), FromServiceHandlerToHandler(sync))(key, obj)
	})
}

func (c *serviceController) Enqueue(namespace, name string) {
	c.controller.Enqueue(namespace, name)
}

func (c *serviceController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controller.EnqueueAfter(namespace, name, duration)
}

func (c *serviceController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *serviceController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *serviceController) Cache() ServiceCache {
	return &serviceCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *serviceController) Create(obj *v1.Service) (*v1.Service, error) {
	result := &v1.Service{}
	return result, c.client.Create(context.TODO(), obj.Namespace, obj, result, metav1.CreateOptions{})
}

func (c *serviceController) Update(obj *v1.Service) (*v1.Service, error) {
	result := &v1.Service{}
	return result, c.client.Update(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *serviceController) UpdateStatus(obj *v1.Service) (*v1.Service, error) {
	result := &v1.Service{}
	return result, c.client.UpdateStatus(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *serviceController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), namespace, name, *options)
}

func (c *serviceController) Get(namespace, name string, options metav1.GetOptions) (*v1.Service, error) {
	result := &v1.Service{}
	return result, c.client.Get(context.TODO(), namespace, name, result, options)
}

func (c *serviceController) List(namespace string, opts metav1.ListOptions) (*v1.ServiceList, error) {
	result := &v1.ServiceList{}
	return result, c.client.List(context.TODO(), namespace, result, opts)
}

func (c *serviceController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), namespace, opts)
}

func (c *serviceController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*v1.Service, error) {
	result := &v1.Service{}
	return result, c.client.Patch(context.TODO(), namespace, name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type serviceCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *serviceCache) Get(namespace, name string) (*v1.Service, error) {
	obj, exists, err := c.indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v1.Service), nil
}

func (c *serviceCache) List(namespace string, selector labels.Selector) (ret []*v1.Service, err error) {

	err = cache.ListAllByNamespace(c.indexer, namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Service))
	})

	return ret, err
}

func (c *serviceCache) AddIndexer(indexName string, indexer ServiceIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.Service))
		},
	}))
}

func (c *serviceCache) GetByIndex(indexName, key string) (result []*v1.Service, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1.Service, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1.Service))
	}
	return result, nil
}

type ServiceStatusHandler func(obj *v1.Service, status v1.ServiceStatus) (v1.ServiceStatus, error)

type ServiceGeneratingHandler func(obj *v1.Service, status v1.ServiceStatus) ([]runtime.Object, v1.ServiceStatus, error)

func RegisterServiceStatusHandler(ctx context.Context, controller ServiceController, condition condition.Cond, name string, handler ServiceStatusHandler) {
	statusHandler := &serviceStatusHandler{
		client:    controller,
		condition: condition,
		handler: func(obj *v1.Service, status v1.ServiceStatus) (v1.ServiceStatus, error) {
			grapher.Record(grapher.Event{Kind: "Handler called", GVK: controller.GroupVersionKind().String(), Key: "status", Name: name, Function: grapher.HandlerFuncName(handler)})
			return handler(obj, status)
		},
	}
	controller.AddGenericHandler(ctx, name, FromServiceHandlerToHandler(statusHandler.sync))
}

func RegisterServiceGeneratingHandler(ctx context.Context, controller ServiceController, apply apply.Apply,
	condition condition.Cond, name string, handler ServiceGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &serviceGeneratingHandler{
		ServiceGeneratingHandler: func(obj *v1.Service, status v1.ServiceStatus) ([]runtime.Object, v1.ServiceStatus, error) {
			grapher.Record(grapher.Event{Kind: "Handler called", GVK: controller.GroupVersionKind().String(), Key: "generating", Name: name, Function: grapher.HandlerFuncName(handler)})
			return handler(obj, status)
		},
		apply: apply,
		name:  name,
		gvk:   controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterServiceStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type serviceStatusHandler struct {
	client    ServiceClient
	condition condition.Cond
	handler   ServiceStatusHandler
}

func (a *serviceStatusHandler) sync(key string, obj *v1.Service) (*v1.Service, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type serviceGeneratingHandler struct {
	ServiceGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *serviceGeneratingHandler) Remove(key string, obj *v1.Service) (*v1.Service, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1.Service{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *serviceGeneratingHandler) Handle(obj *v1.Service, status v1.ServiceStatus) (v1.ServiceStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.ServiceGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
