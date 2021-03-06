/*
Copyright 2017 The Kubernetes Authors.

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

package server

import (
	"fmt"

	"golang.org/x/net/context"
	healthapi "google.golang.org/grpc/health/grpc_health_v1"

	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

const (
	// runtimeNotReadyReason is the reason reported when runtime is not ready.
	runtimeNotReadyReason = "ContainerdNotReady"
	// networkNotReadyReason is the reason reported when network is not ready.
	networkNotReadyReason = "NetworkPluginNotReady"
)

// Status returns the status of the runtime.
func (c *criContainerdService) Status(ctx context.Context, r *runtime.StatusRequest) (*runtime.StatusResponse, error) {
	runtimeCondition := &runtime.RuntimeCondition{
		Type:   runtime.RuntimeReady,
		Status: true,
	}
	// Use containerd grpc server healthcheck service to check its readiness.
	resp, err := c.healthService.Check(ctx, &healthapi.HealthCheckRequest{})
	if err != nil || resp.Status != healthapi.HealthCheckResponse_SERVING {
		runtimeCondition.Status = false
		runtimeCondition.Reason = runtimeNotReadyReason
		if err != nil {
			runtimeCondition.Message = fmt.Sprintf("Containerd healthcheck returns error: %v", err)
		} else {
			runtimeCondition.Message = "Containerd grpc server is not serving"
		}
	}

	networkCondition := &runtime.RuntimeCondition{
		Type:   runtime.NetworkReady,
		Status: true,
	}
	if err := c.netPlugin.Status(); err != nil {
		networkCondition.Status = false
		networkCondition.Reason = networkNotReadyReason
		networkCondition.Message = fmt.Sprintf("Network plugin returns error: %v", err)
	}
	return &runtime.StatusResponse{
		Status: &runtime.RuntimeStatus{Conditions: []*runtime.RuntimeCondition{
			runtimeCondition,
			networkCondition,
		}},
	}, nil
}
