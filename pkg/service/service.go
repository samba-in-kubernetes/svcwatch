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

// Package service implements service-watch probe
package service

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	sf "github.com/samba-in-kubernetes/svcwatch/pkg/statefile"
)

// ToHostState converts a Service type to a HostState.
func ToHostState(svc *corev1.Service, nameLabel string) sf.HostState {
	name := svc.Labels[nameLabel]
	if name == "" {
		// name label not found, fall back to object name
		name = svc.Name
	}
	hs := sf.HostState{
		Reference: fmt.Sprintf("k8s: %s service/%s", svc.Namespace, svc.Name),
		Items:     []sf.HostInfo{},
	}
	for i, ig := range svc.Status.LoadBalancer.Ingress {
		var n string
		if i == 0 {
			n = name
		} else {
			n = fmt.Sprintf("%s-%d", name, i)
		}
		hs.Items = append(hs.Items, sf.HostInfo{
			Name:        n,
			IPv4Address: ig.IP,
			Target:      "external",
		})
	}
	if svc.Spec.ClusterIP != "" {
		hs.Items = append(hs.Items, sf.HostInfo{
			Name:        fmt.Sprintf("%s-cluster", name),
			IPv4Address: svc.Spec.ClusterIP,
			Target:      "internal",
		})
	}
	return hs
}

// Updated detects if a service has been updated by comparing it to
// a previous HostState. It returns the service converted to HostState
// and a boolean indicating if the service has been updated.
func Updated(prev sf.HostState, svc *corev1.Service, nameLabel string) (
	sf.HostState, bool) {
	// ---
	newh := ToHostState(svc, nameLabel)
	return newh, prev.Differs(newh)
}
