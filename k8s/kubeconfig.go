/*
Copyright 2022 The Kubernetes Authors.

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

package k8s

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/cluster"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	configclient "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/anngdinh/operator-helper/contexts"
)

type Getter interface {
	GetClusterKubeconfig(ctx context.Context, clusterKey ctrlclient.ObjectKey) (string, error)
}

type KubeconfigGetter struct{}

// GetClusterKubeconfig returns the kubeconfig for a selected Cluster as a string.
func (k *KubeconfigGetter) GetClusterKubeconfig(ctx context.Context, clusterKey ctrlclient.ObjectKey) (string, error) {
	// logger := contexts.NewContext(ctx).Log()

	// logger.Info("Initializing management cluster kubeconfig")
	managementKubeconfig, err := initInClusterKubeconfig(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "failed to initialize management cluster kubeconfig")
	}

	c, err := client.New(ctx, "")
	if err != nil {
		return "", err
	}

	options := client.GetKubeconfigOptions{
		Kubeconfig:          client.Kubeconfig(*managementKubeconfig),
		WorkloadClusterName: clusterKey.Name,
		Namespace:           clusterKey.Namespace,
	}

	// logger.Infof("Getting kubeconfig for cluster %s.", clusterKey.Name)
	kubeconfig, err := c.GetKubeconfig(ctx, options)
	if err != nil {
		return "", err
	}

	return kubeconfig, nil
}

// initInClusterKubeconfig generates a kubeconfig file for the management cluster.
// Note: The k8s.io/client-go/tools/clientcmd/api package and associated tools require a path to a kubeconfig file rather than the data stored in an object.
func initInClusterKubeconfig(ctx context.Context) (*cluster.Kubeconfig, error) {
	logger := contexts.NewContext(ctx).Log()

	// logger.Info("Generating kubeconfig file")
	restConfig := configclient.GetConfigOrDie()

	apiConfig := constructInClusterKubeconfig(ctx, restConfig, "")

	filePath := "tmp/management.kubeconfig"
	if err := writeInClusterKubeconfigToFile(ctx, filePath, *apiConfig); err != nil {
		logger.Error(err, "error writing kubeconfig to file")
		return nil, err
	}
	kubeconfigPath := filePath
	kubeContext := apiConfig.CurrentContext

	return &cluster.Kubeconfig{Path: kubeconfigPath, Context: kubeContext}, nil
}

// GetClusterKubeconfig generates a kubeconfig file for the management cluster using a rest.Config. This is a bit of a workaround
// since the k8s.io/client-go/tools/clientcmd/api expects to be run from a CLI context, but within a pod we don't have that.
// As a result, we have to manually fill in the fields that would normally be present in ~/.kube/config. This seems to work for now.
func constructInClusterKubeconfig(ctx context.Context, restConfig *rest.Config, namespace string) *clientcmdapi.Config {
	// logger := contexts.NewContext(ctx).Log()

	// logger.Info("Constructing kubeconfig file from rest.Config")

	clusterName := "management-cluster"
	userName := "default-user"
	contextName := "default-context"
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters[clusterName] = &clientcmdapi.Cluster{
		Server: restConfig.Host,
		// Used in regular kubeconfigs.
		CertificateAuthorityData: restConfig.CAData,
		// Used in in-cluster configs.
		CertificateAuthority: restConfig.CAFile,
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts[contextName] = &clientcmdapi.Context{
		Cluster:   clusterName,
		Namespace: namespace,
		AuthInfo:  userName,
	}

	authInfos := make(map[string]*clientcmdapi.AuthInfo)
	authInfos[userName] = &clientcmdapi.AuthInfo{
		Token:                 restConfig.BearerToken,
		ClientCertificateData: restConfig.TLSClientConfig.CertData,
		ClientKeyData:         restConfig.TLSClientConfig.KeyData,
	}

	return &clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: contextName,
		AuthInfos:      authInfos,
	}
}

// writeInClusterKubeconfigToFile writes the clientcmdapi.Config to a kubeconfig file.
func writeInClusterKubeconfigToFile(ctx context.Context, filePath string, clientConfig clientcmdapi.Config) error {
	// logger := contexts.NewContext(ctx).Log()

	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "failed to create directory %s", dir)
		}
	}

	// logger.Infof("Writing kubeconfig to location %s.", filePath)
	if err := clientcmd.WriteToFile(clientConfig, filePath); err != nil {
		return err
	}

	return nil
}
