package resourcebuilder

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/openshift/cluster-version-operator/pkg/payload"
)

func (b *builder) modifyDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	// if proxy injection is requested, get the proxy values and use them
	if containerNamesString := deployment.Annotations["config.openshift.io/inject-proxy"]; len(
		containerNamesString,
	) > 0 {
		proxyConfig, err := b.configClientv1.Proxies().Get(ctx, "cluster", metav1.GetOptions{})
		// not found just means that we don't have proxy configuration, so we should tolerate and fill in empty
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		err = updatePodSpecWithProxy(
			&deployment.Spec.Template.Spec,
			strings.Split(containerNamesString, ","),
			proxyConfig.Status.HTTPProxy,
			proxyConfig.Status.HTTPSProxy,
			proxyConfig.Status.NoProxy,
		)
		if err != nil {
			return err
		}
	}

	// if we detect the CVO deployment we need to replace the KUBERNETES_SERVICE_HOST env var with the internal load
	// balancer to be resilient to kube-apiserver rollouts that cause the localhost server to become non-responsive for
	// multiple minutes.
	if deployment.Namespace == "openshift-cluster-version" && deployment.Name == "cluster-version-operator" {
		infrastructureConfig, err := b.configClientv1.Infrastructures().Get(ctx, "cluster", metav1.GetOptions{})
		// not found just means that we don't have infrastructure configuration yet, so we should tolerate not found and avoid substitution
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if !errors.IsNotFound(err) {
			lbURL, err := url.Parse(infrastructureConfig.Status.APIServerInternalURL)
			if err != nil {
				return err
			}
			// if we have any error and have empty strings, substitution below will do nothing and leave the manifest specified value
			// errors can happen when the port is not specified, in which case we have a host and we write that into the env vars
			lbHost, lbPort, err := net.SplitHostPort(lbURL.Host)
			if err != nil {
				if strings.Contains(err.Error(), "missing port in address") {
					lbHost = lbURL.Host
					lbPort = ""
				} else {
					return err
				}
			}
			err = updatePodSpecWithInternalLoadBalancerKubeService(
				&deployment.Spec.Template.Spec,
				[]string{"cluster-version-operator"},
				lbHost,
				lbPort,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *builder) checkDeploymentHealth(ctx context.Context, deployment *appsv1.Deployment) error {
	if b.mode == InitializingMode {
		return nil
	}

	iden := fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name)
	d, err := b.appsClientv1.Deployments(deployment.Namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if d.DeletionTimestamp != nil {
		return fmt.Errorf("deployment %s is being deleted", iden)
	}

	var availableCondition *appsv1.DeploymentCondition
	var progressingCondition *appsv1.DeploymentCondition
	var replicaFailureCondition *appsv1.DeploymentCondition
	for idx, dc := range d.Status.Conditions {
		switch dc.Type {
		case appsv1.DeploymentProgressing:
			progressingCondition = &d.Status.Conditions[idx]
		case appsv1.DeploymentAvailable:
			availableCondition = &d.Status.Conditions[idx]
		case appsv1.DeploymentReplicaFailure:
			replicaFailureCondition = &d.Status.Conditions[idx]
		}
	}

	if replicaFailureCondition != nil && replicaFailureCondition.Status == corev1.ConditionTrue {
		return &payload.UpdateError{
			Nested: fmt.Errorf(
				"deployment %s has some pods failing; unavailable replicas=%d",
				iden,
				d.Status.UnavailableReplicas,
			),
			Reason: "WorkloadNotProgressing",
			Message: fmt.Sprintf(
				"deployment %s has a replica failure %s: %s",
				iden,
				replicaFailureCondition.Reason,
				replicaFailureCondition.Message,
			),
			Name: iden,
		}
	}

	if availableCondition != nil && availableCondition.Status == corev1.ConditionFalse && progressingCondition != nil &&
		progressingCondition.Status == corev1.ConditionFalse {
		return &payload.UpdateError{
			Nested: fmt.Errorf(
				"deployment %s is not available and not progressing; updated replicas=%d of %d, available replicas=%d of %d",
				iden,
				d.Status.UpdatedReplicas,
				d.Status.Replicas,
				d.Status.AvailableReplicas,
				d.Status.Replicas,
			),
			Reason: "WorkloadNotAvailable",
			Message: fmt.Sprintf(
				"deployment %s is not available %s (%s) or progressing %s (%s)",
				iden,
				availableCondition.Reason,
				availableCondition.Message,
				progressingCondition.Reason,
				progressingCondition.Message,
			),
			Name: iden,
		}
	}

	if availableCondition == nil && progressingCondition == nil && replicaFailureCondition == nil {
		klog.Warningf(
			"deployment %s is not setting any expected conditions, and is therefore in an unknown state",
			iden,
		)
	}

	return nil
}

func (b *builder) modifyDaemonSet(ctx context.Context, daemonset *appsv1.DaemonSet) error {
	// if proxy injection is requested, get the proxy values and use them
	if containerNamesString := daemonset.Annotations["config.openshift.io/inject-proxy"]; len(
		containerNamesString,
	) > 0 {
		proxyConfig, err := b.configClientv1.Proxies().Get(ctx, "cluster", metav1.GetOptions{})
		// not found just means that we don't have proxy configuration, so we should tolerate and fill in empty
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		err = updatePodSpecWithProxy(
			&daemonset.Spec.Template.Spec,
			strings.Split(containerNamesString, ","),
			proxyConfig.Status.HTTPProxy,
			proxyConfig.Status.HTTPSProxy,
			proxyConfig.Status.NoProxy,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *builder) checkDaemonSetHealth(ctx context.Context, daemonset *appsv1.DaemonSet) error {
	if b.mode == InitializingMode {
		return nil
	}

	iden := fmt.Sprintf("%s/%s", daemonset.Namespace, daemonset.Name)
	d, err := b.appsClientv1.DaemonSets(daemonset.Namespace).Get(ctx, daemonset.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if d.DeletionTimestamp != nil {
		return fmt.Errorf("daemonset %s is being deleted", iden)
	}

	// Kubernetes DaemonSet controller doesn't set status conditions yet (v1.18.0), so nothing more to check.
	return nil
}
