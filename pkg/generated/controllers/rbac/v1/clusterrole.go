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
	v1 "k8s.io/api/rbac/v1"
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

type ClusterRoleHandler func(string, *v1.ClusterRole) (*v1.ClusterRole, error)

type ClusterRoleController interface {
	generic.ControllerMeta
	ClusterRoleClient

	OnChange(ctx context.Context, name string, sync ClusterRoleHandler)
	OnRemove(ctx context.Context, name string, sync ClusterRoleHandler)
	Enqueue(name string)
	EnqueueAfter(name string, duration time.Duration)

	Cache() ClusterRoleCache
}

type ClusterRoleClient interface {
	Create(*v1.ClusterRole) (*v1.ClusterRole, error)
	Update(*v1.ClusterRole) (*v1.ClusterRole, error)

	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string, options metav1.GetOptions) (*v1.ClusterRole, error)
	List(opts metav1.ListOptions) (*v1.ClusterRoleList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ClusterRole, err error)
}

type ClusterRoleCache interface {
	Get(name string) (*v1.ClusterRole, error)
	List(selector labels.Selector) ([]*v1.ClusterRole, error)

	AddIndexer(indexName string, indexer ClusterRoleIndexer)
	GetByIndex(indexName, key string) ([]*v1.ClusterRole, error)
}

type ClusterRoleIndexer func(obj *v1.ClusterRole) ([]string, error)

type clusterRoleController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewClusterRoleController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) ClusterRoleController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &clusterRoleController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromClusterRoleHandlerToHandler(sync ClusterRoleHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.ClusterRole
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.ClusterRole))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *clusterRoleController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.ClusterRole))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateClusterRoleDeepCopyOnChange(client ClusterRoleClient, obj *v1.ClusterRole, handler func(obj *v1.ClusterRole) (*v1.ClusterRole, error)) (*v1.ClusterRole, error) {
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

func (c *clusterRoleController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *clusterRoleController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(handler)})
		return generic.NewRemoveHandler(name, c.Updater(), handler)(key, obj)
	})
}

func (c *clusterRoleController) OnChange(ctx context.Context, name string, sync ClusterRoleHandler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(sync)})
		return FromClusterRoleHandlerToHandler(sync)(key, obj)
	})
}

func (c *clusterRoleController) OnRemove(ctx context.Context, name string, sync ClusterRoleHandler) {
	c.AddGenericHandler(ctx, name, func(key string, obj runtime.Object) (runtime.Object, error) {
		grapher.Record(grapher.Event{Kind: "Handler called", GVK: c.gvk.String(), Key: key, Name: name, Function: grapher.HandlerFuncName(sync)})
		return generic.NewRemoveHandler(name, c.Updater(), FromClusterRoleHandlerToHandler(sync))(key, obj)
	})
}

func (c *clusterRoleController) Enqueue(name string) {
	c.controller.Enqueue("", name)
}

func (c *clusterRoleController) EnqueueAfter(name string, duration time.Duration) {
	c.controller.EnqueueAfter("", name, duration)
}

func (c *clusterRoleController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *clusterRoleController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *clusterRoleController) Cache() ClusterRoleCache {
	return &clusterRoleCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *clusterRoleController) Create(obj *v1.ClusterRole) (*v1.ClusterRole, error) {
	result := &v1.ClusterRole{}
	return result, c.client.Create(context.TODO(), "", obj, result, metav1.CreateOptions{})
}

func (c *clusterRoleController) Update(obj *v1.ClusterRole) (*v1.ClusterRole, error) {
	result := &v1.ClusterRole{}
	return result, c.client.Update(context.TODO(), "", obj, result, metav1.UpdateOptions{})
}

func (c *clusterRoleController) Delete(name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), "", name, *options)
}

func (c *clusterRoleController) Get(name string, options metav1.GetOptions) (*v1.ClusterRole, error) {
	result := &v1.ClusterRole{}
	return result, c.client.Get(context.TODO(), "", name, result, options)
}

func (c *clusterRoleController) List(opts metav1.ListOptions) (*v1.ClusterRoleList, error) {
	result := &v1.ClusterRoleList{}
	return result, c.client.List(context.TODO(), "", result, opts)
}

func (c *clusterRoleController) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), "", opts)
}

func (c *clusterRoleController) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1.ClusterRole, error) {
	result := &v1.ClusterRole{}
	return result, c.client.Patch(context.TODO(), "", name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type clusterRoleCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *clusterRoleCache) Get(name string) (*v1.ClusterRole, error) {
	obj, exists, err := c.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v1.ClusterRole), nil
}

func (c *clusterRoleCache) List(selector labels.Selector) (ret []*v1.ClusterRole, err error) {

	err = cache.ListAll(c.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ClusterRole))
	})

	return ret, err
}

func (c *clusterRoleCache) AddIndexer(indexName string, indexer ClusterRoleIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.ClusterRole))
		},
	}))
}

func (c *clusterRoleCache) GetByIndex(indexName, key string) (result []*v1.ClusterRole, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1.ClusterRole, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1.ClusterRole))
	}
	return result, nil
}
