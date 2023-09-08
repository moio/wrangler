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

type PersistentVolumeHandler func(string, *v1.PersistentVolume) (*v1.PersistentVolume, error)

type PersistentVolumeController interface {
	generic.ControllerMeta
	PersistentVolumeClient

	OnChange(ctx context.Context, name string, sync PersistentVolumeHandler)
	OnRemove(ctx context.Context, name string, sync PersistentVolumeHandler)
	Enqueue(name string)
	EnqueueAfter(name string, duration time.Duration)

	Cache() PersistentVolumeCache
}

type PersistentVolumeClient interface {
	Create(*v1.PersistentVolume) (*v1.PersistentVolume, error)
	Update(*v1.PersistentVolume) (*v1.PersistentVolume, error)
	UpdateStatus(*v1.PersistentVolume) (*v1.PersistentVolume, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string, options metav1.GetOptions) (*v1.PersistentVolume, error)
	List(opts metav1.ListOptions) (*v1.PersistentVolumeList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.PersistentVolume, err error)
}

type PersistentVolumeCache interface {
	Get(name string) (*v1.PersistentVolume, error)
	List(selector labels.Selector) ([]*v1.PersistentVolume, error)

	AddIndexer(indexName string, indexer PersistentVolumeIndexer)
	GetByIndex(indexName, key string) ([]*v1.PersistentVolume, error)
}

type PersistentVolumeIndexer func(obj *v1.PersistentVolume) ([]string, error)

type persistentVolumeController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewPersistentVolumeController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) PersistentVolumeController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &persistentVolumeController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromPersistentVolumeHandlerToHandler(sync PersistentVolumeHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.PersistentVolume
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.PersistentVolume))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *persistentVolumeController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.PersistentVolume))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdatePersistentVolumeDeepCopyOnChange(client PersistentVolumeClient, obj *v1.PersistentVolume, handler func(obj *v1.PersistentVolume) (*v1.PersistentVolume, error)) (*v1.PersistentVolume, error) {
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

func (c *persistentVolumeController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *persistentVolumeController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(handler)})
		return generic.NewRemoveHandler(name, c.Updater(), handler)(key, obj)
	})
}

func (c *persistentVolumeController) OnChange(ctx context.Context, name string, sync PersistentVolumeHandler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(sync)})
		return FromPersistentVolumeHandlerToHandler(sync)(key, obj)
	})
}

func (c *persistentVolumeController) OnRemove(ctx context.Context, name string, sync PersistentVolumeHandler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(sync)})
		return generic.NewRemoveHandler(name, c.Updater(), FromPersistentVolumeHandlerToHandler(sync))(key, obj)
	})
}

func (c *persistentVolumeController) Enqueue(name string) {
	c.controller.Enqueue("", name)
}

func (c *persistentVolumeController) EnqueueAfter(name string, duration time.Duration) {
	c.controller.EnqueueAfter("", name, duration)
}

func (c *persistentVolumeController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *persistentVolumeController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *persistentVolumeController) Cache() PersistentVolumeCache {
	return &persistentVolumeCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *persistentVolumeController) Create(obj *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	result := &v1.PersistentVolume{}
	return result, c.client.Create(context.TODO(), "", obj, result, metav1.CreateOptions{})
}

func (c *persistentVolumeController) Update(obj *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	result := &v1.PersistentVolume{}
	return result, c.client.Update(context.TODO(), "", obj, result, metav1.UpdateOptions{})
}

func (c *persistentVolumeController) UpdateStatus(obj *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	result := &v1.PersistentVolume{}
	return result, c.client.UpdateStatus(context.TODO(), "", obj, result, metav1.UpdateOptions{})
}

func (c *persistentVolumeController) Delete(name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), "", name, *options)
}

func (c *persistentVolumeController) Get(name string, options metav1.GetOptions) (*v1.PersistentVolume, error) {
	result := &v1.PersistentVolume{}
	return result, c.client.Get(context.TODO(), "", name, result, options)
}

func (c *persistentVolumeController) List(opts metav1.ListOptions) (*v1.PersistentVolumeList, error) {
	result := &v1.PersistentVolumeList{}
	return result, c.client.List(context.TODO(), "", result, opts)
}

func (c *persistentVolumeController) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), "", opts)
}

func (c *persistentVolumeController) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1.PersistentVolume, error) {
	result := &v1.PersistentVolume{}
	return result, c.client.Patch(context.TODO(), "", name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type persistentVolumeCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *persistentVolumeCache) Get(name string) (*v1.PersistentVolume, error) {
	obj, exists, err := c.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v1.PersistentVolume), nil
}

func (c *persistentVolumeCache) List(selector labels.Selector) (ret []*v1.PersistentVolume, err error) {

	err = cache.ListAll(c.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.PersistentVolume))
	})

	return ret, err
}

func (c *persistentVolumeCache) AddIndexer(indexName string, indexer PersistentVolumeIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.PersistentVolume))
		},
	}))
}

func (c *persistentVolumeCache) GetByIndex(indexName, key string) (result []*v1.PersistentVolume, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1.PersistentVolume, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1.PersistentVolume))
	}
	return result, nil
}

type PersistentVolumeStatusHandler func(obj *v1.PersistentVolume, status v1.PersistentVolumeStatus) (v1.PersistentVolumeStatus, error)

type PersistentVolumeGeneratingHandler func(obj *v1.PersistentVolume, status v1.PersistentVolumeStatus) ([]runtime.Object, v1.PersistentVolumeStatus, error)

func RegisterPersistentVolumeStatusHandler(ctx context.Context, controller PersistentVolumeController, condition condition.Cond, name string, handler PersistentVolumeStatusHandler) {
	statusHandler := &persistentVolumeStatusHandler{
		client:    controller,
		condition: condition,
		handler: func(obj *v1.PersistentVolume, status v1.PersistentVolumeStatus) (v1.PersistentVolumeStatus, error) {
			grapher.Record(grapher.Event{Kind: "Handler called", GVK: controller.GroupVersionKind().String(), Key: "status", Name: name, Function: grapher.HandlerFuncName(handler)})
			return handler(obj, status)
		},
	}
	controller.AddGenericHandler(ctx, name, FromPersistentVolumeHandlerToHandler(statusHandler.sync))
}

func RegisterPersistentVolumeGeneratingHandler(ctx context.Context, controller PersistentVolumeController, apply apply.Apply,
	condition condition.Cond, name string, handler PersistentVolumeGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &persistentVolumeGeneratingHandler{
		PersistentVolumeGeneratingHandler: func(obj *v1.PersistentVolume, status v1.PersistentVolumeStatus) ([]runtime.Object, v1.PersistentVolumeStatus, error) {
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
	RegisterPersistentVolumeStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type persistentVolumeStatusHandler struct {
	client    PersistentVolumeClient
	condition condition.Cond
	handler   PersistentVolumeStatusHandler
}

func (a *persistentVolumeStatusHandler) sync(key string, obj *v1.PersistentVolume) (*v1.PersistentVolume, error) {
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

type persistentVolumeGeneratingHandler struct {
	PersistentVolumeGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *persistentVolumeGeneratingHandler) Remove(key string, obj *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1.PersistentVolume{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *persistentVolumeGeneratingHandler) Handle(obj *v1.PersistentVolume, status v1.PersistentVolumeStatus) (v1.PersistentVolumeStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.PersistentVolumeGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
