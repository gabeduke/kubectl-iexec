package iexec

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
)

// get all pods from kubernetes API
func (r *Iexec) getAllPods() (*corev1.PodList, error) {
	pods, err := r.client.CoreV1().Pods(r.config.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return pods, err
	}

	log.WithFields(log.Fields{
		"pods":      len(pods.Items),
		"namespace": r.config.Namespace,
	}).Debug("total pods discovered...")

	return pods, nil

}

// get all containers for pod
func (r *Iexec) getAllContainers(pod v1.Pod) ([]string, error) {
	log.WithFields(log.Fields{
		"PodName": pod.GetName(),
	}).Infof("Get all containers for pod...")
	var s []string

	containers, err := r.client.CoreV1().Pods(pod.GetNamespace()).Get(pod.GetName(), metav1.GetOptions{})
	log.Tracef("%+v\n", containers)
	if err != nil {
		return nil, err
	}

	l := containers.Spec.Containers
	log.Tracef("%+v\n", l)

	for i, c := range l {
		log.WithFields(log.Fields{
			"index":         i,
			"ContainerName": c.Name,
		}).Infof("Found container...")
		s = append(s, c.Name)
	}

	log.WithFields(log.Fields{
		"[]containers": s,
	}).Debugf("containers []String")
	return s, nil

}

func (r *Iexec) matchPods(pods *corev1.PodList) ([]v1.Pod, error) {
	log.WithFields(log.Fields{
		"SearchFilter": r.search,
	}).Infof("Get all pods for search filter...")
	var matchingPods []v1.Pod

	for i, pod := range pods.Items {
		if strings.Contains(pod.GetName(), r.search) {
			matchingPods = append(matchingPods, pod)
			log.WithFields(log.Fields{
				"PodName": pod.GetName(),
				"index":   i,
			}).Infof("Found pod...")
		}
	}

	if len(matchingPods) == 0 {
		err := fmt.Errorf("No pods found for filter: %s", r.search)
		return nil, err
	}

	sort.Slice(matchingPods[:], func(i, j int) bool {
		return matchingPods[i].GetName() < matchingPods[j].GetName()
	})

	return matchingPods, nil
}

