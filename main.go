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

// Executes svcwatch process.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	goruntime "runtime"
	"syscall"

	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/samba-in-kubernetes/svcwatch/pkg/service"
	sf "github.com/samba-in-kubernetes/svcwatch/pkg/statefile"
)

var (
	// Version of the software at compile time.
	Version = "(unset)"
	// CommitID in the git revision used at compile time.
	CommitID = "(unset)"

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
		os.Exit(1)
	}

	l.Info("Initializing service watcher",
		zap.Any("ProgramName", os.Args[0]),
		zap.Any("Version", Version),
		zap.Any("CommitID", CommitID),
		zap.Any("GoVersion", goruntime.Version()),
	)

	l.Info("Starting service watcher. Params:",
		zap.Any("destination path", destPath),
		zap.Any("label key", svcLabelKey),
		zap.Any("label value to match", svcLabelValue),
		zap.Any("svcNamespace", svcNamespace),
		zap.Any("KUBECONFIG", os.Getenv("KUBECONFIG")),
	)

	clientset, err := newClientset()
	if err != nil {
		l.Error("failed to create clientset", zap.Error(err))
		os.Exit(2)
	}

	sel := fmt.Sprintf("%s=%s", svcLabelKey, svcLabelValue)
	w, err := clientset.CoreV1().Services(svcNamespace).Watch(
		context.TODO(),
		metav1.ListOptions{LabelSelector: sel})
	if err != nil {
		l.Error("failed to create watch", zap.Error(err))
		os.Exit(3)
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

func newClientset() (*kubernetes.Clientset, error) {
	cfg, err := getClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func getClusterConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		cfg, err = buildOutOfClusterConfig()
	}
	return cfg, err
}

func buildOutOfClusterConfig() (*rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube/config")
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}
