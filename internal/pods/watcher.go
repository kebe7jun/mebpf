/*
Copyright © 2022 Merbridge Authors

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

package pods

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	kubeinformer "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/merbridge/merbridge/config"
)

type WatcherAction interface {
	Start() error
	Shutdown()
}

type Watcher struct {
	Client          kubernetes.Interface
	CurrentNodeName string
	OnAddFunc       func(obj interface{})
	OnUpdateFunc    func(oldObj, newObj interface{})
	OnDeleteFunc    func(obj interface{})
	Stop            chan struct{}
}

func (w *Watcher) Start() error {
	selectByNode := ""
	if !config.IsKind {
		selectByNode = fields.OneTermEqualSelector("spec.nodeName", w.CurrentNodeName).String()
	}
	kubeInformerFactory := kubeinformer.NewFilteredSharedInformerFactory(
		w.Client, 30*time.Second, metav1.NamespaceAll,
		func(o *metav1.ListOptions) {
			o.FieldSelector = selectByNode
		},
	)

	kubeInformerFactory.Core().V1().Pods().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.OnAddFunc,
		UpdateFunc: w.OnUpdateFunc,
		DeleteFunc: w.OnDeleteFunc,
	})
	kubeInformerFactory.Start(w.Stop)
	return nil
}

func (w *Watcher) Shutdown() {
	close(w.Stop)
}

func NewWatcher(watch Watcher) *Watcher {
	return &Watcher{
		Client:          watch.Client,
		CurrentNodeName: watch.CurrentNodeName,
		OnAddFunc:       watch.OnAddFunc,
		OnUpdateFunc:    watch.OnUpdateFunc,
		OnDeleteFunc:    watch.OnDeleteFunc,
		Stop:            make(chan struct{}),
	}
}
