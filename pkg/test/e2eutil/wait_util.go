// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2eutil

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/operator-framework/operator-sdk/pkg/test"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// WaitForDeployment checks to see if a given deployment has a certain number of available replicas after a specified amount of time
// If the deployment does not have the required number of replicas after 5 * retries seconds, the function returns an error
// This can be used in multiple ways, like verifying that a required resource is ready before trying to use it, or to test
// failure handling, like simulated in SimulatePodFail.
func WaitForDeployment(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, replicas int, retryInterval, timeout time.Duration) error {
	return waitForDeployment(t, kubeclient, namespace, name, replicas, retryInterval, timeout, false)
}

// WaitForOperatorDeployment has the same functionality as WaitForDeployment but will no wait for the deployment if the
// test was run with a locally run operator (--up-local flag)
func WaitForOperatorDeployment(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, replicas int, retryInterval, timeout time.Duration) error {
	return waitForDeployment(t, kubeclient, namespace, name, replicas, retryInterval, timeout, true)
}

func waitForDeployment(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, replicas int, retryInterval, timeout time.Duration, isOperator bool) error {
	if isOperator && test.Global.LocalOperator {
		t.Log("Operator is running locally; skip waitForDeployment")
		return nil
	}
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		deployment, err := kubeclient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s deployment\n", name)
				return false, nil
			}
			return false, err
		}

		if int(deployment.Status.AvailableReplicas) == replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s deployment (%d/%d)\n", name, deployment.Status.AvailableReplicas, replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Deployment available (%d/%d)\n", replicas, replicas)
	return nil
}

// WaitForPod checks to see if a given pod is running after a specified amount of time.
// If the deployment is not running after timeout * retries seconds, the function returns an error
func WaitForPod(t *testing.T, runtimeClient test.FrameworkClient, namespace, name string, retryInterval, timeout time.Duration) error {
	pod := &corev1.Pod{}
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = runtimeClient.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: namespace}, pod)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s pod\n", name)
				return false, nil
			}
			return false, err
		}

		if pod.Status.Phase == corev1.PodRunning {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s pod\n", name)
		return false, nil
	})

	if err != nil {
		return err
	}
	t.Logf("%s Pod available\n", name)
	return nil
}
