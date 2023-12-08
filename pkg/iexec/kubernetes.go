package iexec

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// get all pods from kubernetes API.
func getAllPods(client kubernetes.Interface, namespace, selector string) (*corev1.PodList, error) {
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{FieldSelector: "status.phase=Running", LabelSelector: selector})
	if err != nil {
		return pods, errors.Wrap(err, "unable to get pods")
	}

	if len(pods.Items) == 0 {
		return pods, errors.New("no running pods found")
	}

	log.WithFields(log.Fields{
		"pods":      len(pods.Items),
		"namespace": namespace,
	}).Debug("total pods discovered...")

	return pods, nil
}

func (r *Iexec) matchPods(pods *corev1.PodList) (corev1.PodList, error) {
	var result corev1.PodList

	log.WithFields(log.Fields{
		"SearchFilter": r.config.PodFilter,
	}).Infof("Get all pods for podFilter...")

	for i, pod := range pods.Items {
		if strings.Contains(pod.GetName(), r.config.PodFilter) {
			result.Items = append(result.Items, pod)
			log.WithFields(log.Fields{
				"PodName": pod.GetName(),
				"index":   i,
			}).Infof("Found pod...")
		}
	}

	if len(result.Items) == 0 {
		err := fmt.Errorf("no pods found for filter: %s", r.config.PodFilter)

		return result, err
	}

	return result, nil
}

func matchContainers(pod corev1.Pod, config Config) ([]corev1.Container, error) {
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

	sort.Slice(matchingContainer[:], func(i, j int) bool {
		return matchingContainer[i].Name < matchingContainer[j].Name
	})

	return matchingContainer, nil
}
