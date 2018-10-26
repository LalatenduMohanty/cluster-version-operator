/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	clusterversionv1 "github.com/openshift/cluster-version-operator/pkg/generated/clientset/versioned/typed/config.openshift.io/v1"
	operatorstatusv1 "github.com/openshift/cluster-version-operator/pkg/generated/clientset/versioned/typed/operatorstatus.openshift.io/v1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	ClusterversionV1() clusterversionv1.ClusterversionV1Interface
	// Deprecated: please explicitly pick a version if possible.
	Clusterversion() clusterversionv1.ClusterversionV1Interface
	OperatorstatusV1() operatorstatusv1.OperatorstatusV1Interface
	// Deprecated: please explicitly pick a version if possible.
	Operatorstatus() operatorstatusv1.OperatorstatusV1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	clusterversionV1 *clusterversionv1.ClusterversionV1Client
	operatorstatusV1 *operatorstatusv1.OperatorstatusV1Client
}

// ClusterversionV1 retrieves the ClusterversionV1Client
func (c *Clientset) ClusterversionV1() clusterversionv1.ClusterversionV1Interface {
	return c.clusterversionV1
}

// Deprecated: Clusterversion retrieves the default version of ClusterversionClient.
// Please explicitly pick a version.
func (c *Clientset) Clusterversion() clusterversionv1.ClusterversionV1Interface {
	return c.clusterversionV1
}

// OperatorstatusV1 retrieves the OperatorstatusV1Client
func (c *Clientset) OperatorstatusV1() operatorstatusv1.OperatorstatusV1Interface {
	return c.operatorstatusV1
}

// Deprecated: Operatorstatus retrieves the default version of OperatorstatusClient.
// Please explicitly pick a version.
func (c *Clientset) Operatorstatus() operatorstatusv1.OperatorstatusV1Interface {
	return c.operatorstatusV1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.clusterversionV1, err = clusterversionv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.operatorstatusV1, err = operatorstatusv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.clusterversionV1 = clusterversionv1.NewForConfigOrDie(c)
	cs.operatorstatusV1 = operatorstatusv1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.clusterversionV1 = clusterversionv1.New(c)
	cs.operatorstatusV1 = operatorstatusv1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
