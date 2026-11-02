package iexec

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RealK8sClient struct {
	client kubernetes.Interface
}

func NewRealK8sClient(client kubernetes.Interface) *RealK8sClient {
	return &RealK8sClient{client: client}
}

func (r *RealK8sClient) FetchAllPods(namespace, labelSelector string) (*corev1.PodList, error) {
	return r.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
}

func (r *RealK8sClient) MatchPods(pods *corev1.PodList, config Config) (*corev1.PodList, error) {
	var result corev1.PodList

	log.WithFields(log.Fields{
		"SearchFilter": config.PodFilter,
	}).Infof("Get all pods for podFilter...")

	skipped := 0
	for i, pod := range pods.Items {
		if !strings.Contains(pod.GetName(), config.PodFilter) {
			continue
		}
		if !config.AllPhases && !isExecable(pod) {
			skipped++
			continue
		}
		result.Items = append(result.Items, pod)
		log.WithFields(log.Fields{
			"PodName": pod.GetName(),
			"index":   i,
		}).Infof("Found pod...")
	}

	if len(result.Items) == 0 {
		if skipped > 0 {
			return &result, fmt.Errorf("no running pods found for filter: %s (%d not running, use --all-phases to include them)", config.PodFilter, skipped)
		}
		return &result, fmt.Errorf("no pods found for filter: %s", config.PodFilter)
	}

	return &result, nil
}

// isExecable reports whether a pod can accept an exec: Running and not terminating.
func isExecable(pod corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning && pod.DeletionTimestamp == nil
}

func (r *RealK8sClient) MatchContainers(pod corev1.Pod, config Config) ([]corev1.Container, error) {
	if config.ContainerFilter == "" {
		return pod.Spec.Containers, nil
	}
	log.WithFields(log.Fields{
		"SearchFilter": config.ContainerFilter,
	}).Infof("Get all containers for containerFilter...")
	var matchingContainer []corev1.Container

	for i, container := range pod.Spec.Containers {
		if strings.Contains(container.Name, config.ContainerFilter) {
			matchingContainer = append(matchingContainer, container)
			log.WithFields(log.Fields{
				"ContainerName": container.Name,
				"index":         i,
			}).Infof("Found container...")
		}
	}

	if len(matchingContainer) == 0 {
		err := fmt.Errorf("no containers found for filter: %s", config.ContainerFilter)

		return nil, err
	}

	sort.Slice(matchingContainer, func(i, j int) bool {
		return matchingContainer[i].Name < matchingContainer[j].Name
	})

	return matchingContainer, nil
}
