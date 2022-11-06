package watchimpl

import (
	"context"
	"sync"
	"time"

	"github.com/goccy/go-json"
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
	"k8s.io/client-go/kubernetes"
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
	v1.ServerCreated: {},
	v1.ServerReady:   {},
	v1.ServerBound:   {},
}

type Watcher struct {
	res      *kubernetes.Clientset
	resync   time.Duration
	mux      *sync.Mutex
	resource dynamic.NamespaceableResourceInterface
	nConfig  *Config
	stopCh   chan struct{}
}

type StatusDetail struct {
	AccessUrl string `json:"access_url,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

func NewWatcher(cfg *Config) *Watcher {
	return &Watcher{
		res:      k8sclient.GetClient(),
		resource: k8sclient.GetResource(),
		nConfig:  cfg,
	}
}

func (w *Watcher) Run() {
	infor := w.crdConfig()
	infor.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: w.Update,
	})

	w.stopCh = make(chan struct{})

	infor.Run(w.stopCh)

	if !cache.WaitForCacheSync(w.stopCh, infor.HasSynced) {
		logrus.Fatalln("cache sync err")
		return
	}
	<-w.stopCh
}

func (w *Watcher) Update(oldObj, newObj interface{}) {
	var res v1.CodeServer

	bys, err := json.Marshal(newObj)
	if err != nil {
		logrus.Errorf("update marshal error:%s", err.Error())
		return
	}

	err = json.Unmarshal(bys, &res)
	if err != nil {
		logrus.Errorf("update unmarshal error:%s", err.Error())
		return
	}

	go w.dispatcher(res)
}

func (w *Watcher) dispatcher(res v1.CodeServer) {
	status := w.transferStatus(res)

	switch res.Labels["type"] {
	case inferenceimpl.MetaName():
		logrus.Debugf("watched inference crd, status: %s", status)

		w.HandleInference(res.ObjectMeta.Labels, status)

	case evaluateimpl.MetaName():
		logrus.Debugf("watched evaluate crd, status: %s", status)

		w.HandleEvaluate(res.ObjectMeta.Labels, status)
	}
}

func (w *Watcher) transferStatus(res v1.CodeServer) (status StatusDetail) {
	var endPoint string
	for _, condition := range res.Status.Conditions {
		if _, ok := serverUnusable[condition.Type]; ok {
			if condition.Status == corev1.ConditionTrue {
				status.ErrorMsg = condition.Reason
				return
			}
		}

		if _, ok := serverUsable[condition.Type]; ok {
			if condition.Status == corev1.ConditionFalse {
				status.ErrorMsg = condition.Reason
				return
			}
		}
		if endPoint == "" {
			endPoint = condition.Message["instanceEndpoint"]
		}
	}
	status.AccessUrl = endPoint
	return
}

func (w *Watcher) HandleInference(labels map[string]string, status StatusDetail) {
	cli, err := rpcclient.NewInferenceClient(w.nConfig.InferenceEndpoint)
	if err != nil {
		logrus.Errorf("new rpc client error:%s", err.Error())
	}

	index := inference.InferenceIndex{
		Id:         labels["id"],
		User:       labels["user"],
		ProjectId:  labels["project_id"],
		LastCommit: labels["last_commit"],
	}

	info := inference.InferenceInfo{
		Error:     status.ErrorMsg,
		AccessURL: status.AccessUrl,
	}
	if err = cli.SetInferenceInfo(&index, &info); err != nil {
		logrus.Errorf("call inference rpc error:%s", err.Error())
	}

}

func (w *Watcher) HandleEvaluate(labels map[string]string, status StatusDetail) {
	cli, err := rpcclient.NewEvaluateClient(w.nConfig.EvaluateEndpoint)
	if err != nil {
		logrus.Error("new evaluate rpc client error:", err.Error())
	}
	index := evaluate.EvaluateIndex{
		Id:         labels["id"],
		User:       labels["user"],
		ProjectId:  labels["project_id"],
		TrainingID: labels["training_id"],
	}
	info := evaluate.EvaluateInfo{
		Error:     status.ErrorMsg,
		AccessURL: status.AccessUrl,
	}
	if err = cli.SetEvaluateInfo(&index, &info); err != nil {
		logrus.Error("call evaluate rpc error:", err.Error())
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
