package watchimpl

import (
	"context"
	"encoding/json"
	"sync"

	v1 "github.com/opensourceways/code-server-operator/api/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/evaluateimpl"
	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/inferenceimpl"
	"github.com/opensourceways/xihe-inference-evaluate/k8sclient"
)

type Watcher struct {
	cli *k8sclient.Client

	podNamePrifixes []string
	handles         map[string]func(map[string]string, domain.ContainerDetail)
	stop            chan struct{}
	wg              sync.WaitGroup
}

type statusDetail struct {
	accessUrl string
	errorMsg  string
}

func NewWatcher(
	cli *k8sclient.Client,
	handles map[string]func(map[string]string, domain.ContainerDetail),
) *Watcher {
	return &Watcher{
		cli:     cli,
		stop:    make(chan struct{}),
		handles: handles,
		podNamePrifixes: []string{
			inferenceimpl.MetaName(),
			evaluateimpl.MetaName(),
		},
	}
}

func (w *Watcher) Run() {

	w.watchCRD()

	w.wg.Add(1)
	go w.watchPod()
}

func (w *Watcher) watchCRD() {
	infor := w.crdConfig()

	infor.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: w.update,
	})

	w.wg.Add(1)
	go func() {
		logrus.Debug("start watching crd")

		infor.Run(w.stop)

		w.wg.Done()
	}()

	if !cache.WaitForCacheSync(w.stop, infor.HasSynced) {
		logrus.Error("cache sync err")
	} else {
		logrus.Debug("cache sync done")
	}
}

func (w *Watcher) Exit() {
	close(w.stop)

	w.wg.Wait()
}

func (w *Watcher) update(oldObj, newObj interface{}) {
	v, err := json.Marshal(newObj)
	if err != nil {
		logrus.Errorf("update marshal error:%s", err.Error())

		return
	}

	var res v1.CodeServer

	if err = json.Unmarshal(v, &res); err != nil {
		logrus.Errorf("update unmarshal error:%s", err.Error())

		return
	}

	w.wg.Add(1)
	go w.checkCRD(res)
}

func (w *Watcher) checkCRD(res v1.CodeServer) {
	defer w.wg.Done()

	h, ok := w.handles[res.Labels[labelType]]
	if !ok {
		return
	}

	recycled, endPoint := w.checkCRDStatus(&res)

	if recycled {
		/*
			if err := w.cli.DeleteCRD(res.GetName()); err != nil {
				logrus.Errorf("watch delete crd(%s) err: %s", res.GetName(), err.Error())
			}
		*/

		return
	}

	if endPoint != "" {
		h(res.ObjectMeta.Labels, domain.ContainerDetail{AccessUrl: endPoint})

		w.updateCRDBoundStatus(&res)
	}
}

func (w *Watcher) updateCRDBoundStatus(res *v1.CodeServer) {
	bingo := false
	conditions := res.Status.Conditions
	for k := range conditions {
		cond := &conditions[k]

		if cond.Type == v1.ServerBound && cond.Status == corev1.ConditionFalse {
			cond.Status = corev1.ConditionTrue
			cond.Reason = "bind to code server"

			bingo = true
			break
		}
	}

	if !bingo {
		return
	}

	b, err := json.Marshal(&res)
	if err != nil {
		logrus.Errorf("update marshal error:%s", err.Error())

		return
	}

	// unstructured.Unstructured implements the method of UnmarshalJSON
	object := new(unstructured.Unstructured)
	if err = json.Unmarshal(b, object); err != nil {
		logrus.Errorf("update unmarshal error:%s", err.Error())

		return
	}

	if err = w.cli.UpdateCRD(object); err != nil {
		logrus.Errorf("update CRD failed, err:%v", err)
	}
}

func (w *Watcher) checkCRDStatus(res *v1.CodeServer) (recycled bool, endPoint string) {
	v := res.Status.Conditions
	for i := range v {
		item := &v[i]

		switch item.Type {
		case v1.ServerRecycled:
			if item.Status == corev1.ConditionTrue {
				recycled = true

				break
			}

		case v1.ServerReady:
			if item.Status == corev1.ConditionTrue {
				endPoint = item.Message["instanceEndpoint"]

				break
			}
		}
	}

	return
}

func (w *Watcher) crdConfig() cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return w.cli.GetResource().List(context.TODO(), options)
			},

			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return w.cli.GetResource().Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{},
		0,
		cache.Indexers{},
	)
}
