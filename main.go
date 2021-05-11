/*
Copyright 2021 John Mulligan <phlogistonjohn@asynchrono.us>

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

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/samba-in-kubernetes/svcwatch/pkg/service"
	sf "github.com/samba-in-kubernetes/svcwatch/pkg/statefile"
)

var (
	destPath      = "/var/lib/svcwatch/status.json"
	svcLabelKey   = ""
	svcLabelValue = ""
	svcNamespace  = ""
)

func processUpdates(
	path, nameLabel string,
	updates <-chan *corev1.Service, errors chan<- error) {
	// initialize host state to the zero value
	var hs sf.HostState
	changed := true
	for svc := range updates {
		if svc == nil {
			return
		}
		hs, changed = service.Updated(hs, svc, nameLabel)
		if changed {
			err := hs.Save(path)
			if err != nil {
				errors <- err
			}
		}
	}
}

func envFlag(d *string, label, envKey, help string) {
	defVal := os.Getenv(envKey)
	flag.StringVar(
		d, label, defVal,
		fmt.Sprintf("%s (default from: %s)", help, envKey))
}

func main() {
	envFlag(&destPath, "destination", "DESTINATION_PATH", "JSON file to update")
	envFlag(&svcLabelKey, "label-key", "SERVICE_LABEL_KEY", "Label key to watch")
	envFlag(&svcLabelValue, "label-value", "SERVICE_LABEL_VALUE", "Label value")
	envFlag(&svcNamespace, "namespace", "SERVICE_NAMESPACE", "Namespace")
	flag.Parse()

	l, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("failed to set up logger\n")
		os.Exit(2)
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	l.Info("Starting service watcher. Params:",
		zap.Any("destination path", destPath),
		zap.Any("label key", svcLabelKey),
		zap.Any("label value to match", svcLabelValue),
		zap.Any("svcNamespace", svcNamespace),
		zap.Any("KUBECONFIG", kubeconfig),
	)

	kcfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		l.Error("failed to create watch", zap.Error(err))
		os.Exit(1)
	}
	clientset := kubernetes.NewForConfigOrDie(kcfg)

	sel := fmt.Sprintf("%s=%s", svcLabelKey, svcLabelValue)
	w, err := clientset.CoreV1().Services(svcNamespace).Watch(
		context.TODO(),
		metav1.ListOptions{LabelSelector: sel})
	if err != nil {
		l.Error("failed to create watch", zap.Error(err))
		os.Exit(1)
	}

	errors := make(chan error, 1)
	updates := make(chan *corev1.Service)
	signalch := make(chan os.Signal, 1)
	signal.Notify(signalch,
		os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
	ready := true
	go processUpdates(destPath, svcLabelKey, updates, errors)
	for ready {
		select {
		case v := <-w.ResultChan():
			l.Info("new result from watch", zap.Any("value", v))
			svc, ok := v.Object.(*corev1.Service)
			if !ok {
				l.Error("got non-service from service watch")
				ready = false
				break
			}
			l.Info("updated service", zap.Any("service", svc))
			updates <- svc
		case <-signalch:
			l.Info("terminating")
			ready = false
			break
		case e := <-errors:
			l.Error("error updating host state", zap.Error(e))
			ready = false
			break
		}
	}
	updates <- nil
	close(updates)
	close(errors)
	w.Stop()
}
