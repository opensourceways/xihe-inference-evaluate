package watchimpl

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "github.com/opensourceways/code-server-operator/api/v1alpha1"
	rpcclient "github.com/opensourceways/xihe-grpc-protocol/grpc/client"
	"github.com/opensourceways/xihe-grpc-protocol/grpc/evaluate"
	"github.com/opensourceways/xihe-grpc-protocol/grpc/inference"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"

	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/evaluateimpl"
	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/inferenceimpl"
	"github.com/opensourceways/xihe-inference-evaluate/k8sclient"
)

var serverUnusable = map[v1.ServerConditionType]struct{}{
	v1.ServerRecycled: {},
	v1.ServerInactive: {},
	v1.ServerErrored:  {},
}

var serverUsable = map[v1.ServerConditionType]struct{}{
	v1.ServerReady: {},
}

type Watcher struct {
	evaluateClient  *rpcclient.EvaluateClient
	inferenceClient *rpcclient.InferenceClient

	handles  map[string]func(map[string]string, statusDetail)
	resource dynamic.NamespaceableResourceInterface
	stopCh   chan struct{}
}

type statusDetail struct {
	accessUrl string
	errorMsg  string
}

func NewWatcher(cfg *Config) (*Watcher, error) {
	evaluateClient, err := rpcclient.NewEvaluateClient(cfg.EvaluateEndpoint)
	if err != nil {
		return nil, fmt.Errorf("new evaluate rpc client error: %s", err.Error())
	}

	inferenceClient, err := rpcclient.NewInferenceClient(cfg.InferenceEndpoint)
	if err != nil {
		return nil, fmt.Errorf("new evaluate rpc client error: %s", err.Error())
	}

	w := &Watcher{
		resource:        k8sclient.GetResource(),
		evaluateClient:  evaluateClient,
		inferenceClient: inferenceClient,
	}

	w.handles = map[string]func(map[string]string, statusDetail){
		inferenceimpl.MetaName(): w.handleInference,
		evaluateimpl.MetaName():  w.handleEvaluate,
	}

	return w, nil
}

func (w *Watcher) Run() {
	infor := w.crdConfig()
	infor.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: w.update,
	})

	w.stopCh = make(chan struct{})

	infor.Run(w.stopCh)

	if !cache.WaitForCacheSync(w.stopCh, infor.HasSynced) {
		logrus.Fatalln("cache sync err")
		return
	}

	<-w.stopCh
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

	h, ok := w.handles[res.Labels["type"]]
	if !ok {
		return
	}

	go h(res.ObjectMeta.Labels, w.transferStatus(res))
}

func (w *Watcher) transferStatus(res v1.CodeServer) (status statusDetail) {
	var endPoint string

	for _, condition := range res.Status.Conditions {
		if _, ok := serverUnusable[condition.Type]; ok {
			if condition.Status == corev1.ConditionTrue {
				status.errorMsg = condition.Reason
				if msg, ok := condition.Message["detail"]; ok && len(msg) != 0 {
					status.errorMsg = msg
				}

				return
			}
		}

		if _, ok := serverUsable[condition.Type]; ok {
			if condition.Status == corev1.ConditionFalse {
				status.errorMsg = condition.Reason

				return
			}
		}

		if endPoint == "" {
			endPoint = condition.Message["instanceEndpoint"]
		}
	}

	status.accessUrl = endPoint

	return
}

func (w *Watcher) handleInference(labels map[string]string, status statusDetail) {
	index := inference.InferenceIndex{
		Id:         labels["id"],
		User:       labels["user"],
		ProjectId:  labels["project_id"],
		LastCommit: labels["last_commit"],
	}

	info := inference.InferenceInfo{
		Error:     status.errorMsg,
		AccessURL: status.accessUrl,
	}

	err := w.inferenceClient.SetInferenceInfo(&index, &info)
	if err != nil {
		logrus.Errorf("call inference rpc error:%s", err.Error())
	} else {
		logrus.Debugf(
			"call rpc to set inference(%s/%s/%s/%s) info:(%s/%s)",
			index.User, index.ProjectId, index.LastCommit, index.Id,
			info.Error, info.AccessURL,
		)

	}
}

func (w *Watcher) handleEvaluate(labels map[string]string, status statusDetail) {
	index := evaluate.EvaluateIndex{
		Id:         labels["id"],
		User:       labels["user"],
		ProjectId:  labels["project_id"],
		TrainingID: labels["training_id"],
	}
	info := evaluate.EvaluateInfo{
		Error:     status.errorMsg,
		AccessURL: status.accessUrl,
	}

	err := w.evaluateClient.SetEvaluateInfo(&index, &info)
	if err != nil {
		logrus.Errorf("call evaluate rpc error: %s", err.Error())
	} else {
		logrus.Debugf(
			"call rpc to set evaluate(%s/%s/%s/%s) info:(%s/%s)",
			index.User, index.ProjectId, index.TrainingID, index.Id,
			info.Error, info.AccessURL,
		)
	}
}

func (w *Watcher) crdConfig() cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return w.resource.List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return w.resource.Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{},
		0,
		cache.Indexers{},
	)
}

func (w *Watcher) Exit() {
	close(w.stopCh)
}
