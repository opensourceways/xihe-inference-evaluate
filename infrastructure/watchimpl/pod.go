package watchimpl

import (
	"bytes"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

const (
	labelType   = "type"
	labelCSName = "cs_name"
)

func (w *Watcher) watchPod() {
	logrus.Debug("start watching pods.")

	t := time.Tick(time.Second * 2)

	for {
		select {
		case <-t:
			if v, err := w.cli.ListPods(); err != nil {
				logrus.Errorf("list pods failed, err:%s", err.Error())
			} else {
				w.checkPods(v)
			}

		case <-w.stop:
			close(w.podWatchStopped)

			return
		}
	}
}

func (w *Watcher) isTargetPod(pod *corev1.Pod) bool {
	name := pod.GetName()

	for _, s := range w.podNamePrifixes {
		if strings.HasPrefix(name, s) {
			return true
		}
	}

	return false
}

func (w *Watcher) checkPods(pods []corev1.Pod) {
	for i := range pods {
		pod := &pods[i]

		if !w.isTargetPod(pod) || !w.cli.IsPodFailed(pod) {
			continue
		}

		buf := new(bytes.Buffer)

		if err := w.cli.FailedPodLog(pod, buf); err != nil {
			logrus.Errorf(
				"get log of pod(%s) failed, err:%s",
				pod.GetName(), err.Error(),
			)
		}

		labels := w.deleteCRDOfPod(pod)
		if len(labels) == 0 {
			continue
		}

		if h, ok := w.handles[labels[labelType]]; ok {
			d := statusDetail{}

			if buf.Len() == 0 {
				d.errorMsg = "unknown error"
			} else {
				d.errorMsg = buf.String()
			}

			h(labels, d)
		}
	}
}

func (w *Watcher) deleteCRDOfPod(pod *corev1.Pod) (labels map[string]string) {
	name := pod.Labels[labelCSName]

	crd, err := w.cli.GetCRD(name)
	if err != nil {
		logrus.Errorf("get crd resource(%s) err: %s", name, err.Error())

		return
	}

	if err := w.cli.DeleteCRD(name); err != nil {
		logrus.Errorf("get crd resource(%s) err: %s", name, err.Error())
	}

	labels = crd.GetLabels()

	return
}
