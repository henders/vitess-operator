/*
Copyright 2019 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file is forked from github.com/GoogleCloudPlatform/metacontroller.

package framework

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"planetscale.dev/vitess-operator/pkg/operator/controllermanager"
)

const installKubectl = `
Cannot find kubectl, cannot run integration tests

Please download kubectl and ensure it is somewhere in the PATH.
See tools/get-kube-binaries.sh

`

// deployDir is the path from the integration test binary working dir to the
// directory containing manifests to install vitess-operator.
const deployDir = "../../../deploy"

// getKubectlPath returns a path to a kube-apiserver executable.
func getKubectlPath() (string, error) {
	return exec.LookPath("kubectl")
}

// TestMain starts etcd, kube-apiserver, and vitess-operator before running tests.
func TestMain(tests func() int) {
	if err := testMain(tests); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func testMain(tests func() int) error {
	controllermanager.InitFlags()

	if _, err := getKubectlPath(); err != nil {
		return errors.New(installKubectl)
	}

	stopEtcd, err := startEtcd()
	if err != nil {
		return fmt.Errorf("cannot run integration tests: unable to start etcd: %v", err)
	}
	defer stopEtcd()

	stopApiserver, err := startApiserver()
	if err != nil {
		return fmt.Errorf("cannot run integration tests: unable to start kube-apiserver: %v", err)
	}
	defer stopApiserver()

	klog.Info("Waiting for kube-apiserver to be ready...")
	start := time.Now()
	for {
		out, kubectlErr := execKubectl("version")
		if kubectlErr == nil {
			break
		}
		if time.Since(start) > 60*time.Second {
			return fmt.Errorf("timed out waiting for kube-apiserver to be ready: %v\n%s", kubectlErr, out)
		}
		time.Sleep(time.Second)
	}

	// Install vites-operator base files, but not the Deployment itself.
	files := []string{
		"service_account.yaml",
		"role.yaml",
		"role_binding.yaml",
		"priority.yaml",
		"crds/",
	}
	for _, file := range files {
		filePath := path.Join(deployDir, file)
		klog.Infof("Installing %v...", filePath)
		if out, err := execKubectl("apply", "-f", filePath); err != nil {
			return fmt.Errorf("cannot install %v: %v\n%s", filePath, err, out)
		}
	}

	klog.Info("Waiting for CRDs to be ready...")
	start = time.Now()
	for {
		out, kubectlErr := execKubectl("get", "vt,vtc,vtk,vts,vtbs,vtb,etcdls")
		if kubectlErr == nil {
			break
		}
		if time.Since(start) > 30*time.Second {
			return fmt.Errorf("timed out waiting for CRDs to be ready: %v\n%s", kubectlErr, out)
		}
		time.Sleep(time.Second)
	}

	// Start vitess-operator in this test process.
	mgr, err := controllermanager.New("", ApiserverConfig(), manager.Options{
		Namespace: "default",
	})
	if err != nil {
		return fmt.Errorf("cannot create controller-manager: %v", err)
	}
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		if err := mgr.Start(stop); err != nil {
			klog.Errorf("cannot start controller-manager: %v", err)
		}
	}()

	// Now actually run the tests.
	if exitCode := tests(); exitCode != 0 {
		return fmt.Errorf("one or more tests failed with exit code: %v", exitCode)
	}
	return nil
}

func execKubectl(args ...string) ([]byte, error) {
	execPath, err := exec.LookPath("kubectl")
	if err != nil {
		return nil, fmt.Errorf("cannot exec kubectl: %v", err)
	}
	cmdline := append([]string{"--server", ApiserverURL()}, args...)
	cmd := exec.Command(execPath, cmdline...)
	return cmd.CombinedOutput()
}
