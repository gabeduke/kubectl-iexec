package cmd

import (
	"sort"

	"github.com/manifoldco/promptui"

	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	// auth needed for proxy
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// get all pods from kubernetes API
func (r *iexec) getAllPods() (*corev1.PodList, error) {
	pods, err := r.client.CoreV1().Pods(r.namespace).List(metav1.ListOptions{})
	if err != nil {
		return pods, err
	}

	log.WithFields(log.Fields{
		"pods":      len(pods.Items),
		"namespace": r.namespace,
	}).Debug("total pods discovered...")

	return pods, nil

}

// get all containers for pod
func (r *iexec) getAllContainers(pod v1.Pod) ([]string, error) {
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

func (r *iexec) matchPods(pods *corev1.PodList) ([]v1.Pod, error) {
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

func (r *iexec) podPrompt(matchingPods []v1.Pod) error {

	if len(matchingPods) == 1 {
		r.pod = matchingPods[0]
		r.namespace = r.pod.GetNamespace()
		fmt.Printf("Found matching pod (%s) with filter: %s\n", r.pod.Name, r.search)
		return nil
	}

	templates := &promptui.SelectTemplates{
		Active:   fmt.Sprintf("Namespace: {{ .Namespace | blue }} | Pod: %s {{ .Name | cyan }}", promptui.IconSelect),
		Inactive: "Namespace: {{ .Namespace | blue }} | Pod: {{ .Name | magenta }}",
		Selected: fmt.Sprintf("Namespace: {{ .Namespace | blue }} | Pod: %s {{ .Name | cyan }}", promptui.IconGood),
	}

	if naked {
		log.Debugf("naked: %v", naked)
		templates = &promptui.SelectTemplates{
			Active:   fmt.Sprintf("Namespace: {{ .Namespace }} | Pod: %s {{ .Name }}", promptui.IconSelect),
			Inactive: "Namespace: {{ .Namespace }} | Pod: {{ .Name }}",
			Selected: fmt.Sprintf("Namespace: {{ .Namespace }} | Pod: %s {{ .Name }}", promptui.IconGood),
		}

	}

	podsPrompt := promptui.Select{
		Label:     "Select Pod",
		Items:     matchingPods,
		Templates: templates,
		IsVimMode: vimMode,
	}

	i, _, err := podsPrompt.Run()
	if err != nil {
		return err
	}

	r.pod = matchingPods[i]
	r.namespace = r.pod.GetNamespace()
	return nil
}

func (r *iexec) containerPrompt() error {

	containers, err := r.getAllContainers(r.pod)
	if err != nil {
		return err
	}

	if len(containers) == 1 {
		return nil
	}

	templates := &promptui.SelectTemplates{
		Active:   fmt.Sprintf("Container: %s {{ . | cyan }}", promptui.IconSelect),
		Inactive: "Container: {{ . | magenta }}",
		Selected: fmt.Sprintf("Container: %s {{ . | cyan }}", promptui.IconGood),
	}

	if naked {
		log.Debugf("naked: %v", naked)
		templates = &promptui.SelectTemplates{
			Active:   fmt.Sprintf("Container: %s {{ . }}", promptui.IconSelect),
			Inactive: "Container: {{ . }}",
			Selected: fmt.Sprintf("Container: %s {{ . }}", promptui.IconGood),
		}

	}

	containersPrompt := promptui.Select{
		Label:     "Select Container",
		Items:     containers,
		Templates: templates,
		IsVimMode: vimMode,
	}

	c, _, err := containersPrompt.Run()
	if err != nil {
		return err
	}

	r.container = containers[c]

	return nil
}

func (r *iexec) do() error {
	pods, err := r.getAllPods()
	if err != nil {
		return err
	}

	matchingPods, err := r.matchPods(pods)
	if err != nil {
		return err
	}

	err = r.podPrompt(matchingPods)
	if err != nil {
		return err
	}

	err = r.containerPrompt()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"pod":       r.pod.GetName(),
		"container": r.container,
		"namespace": r.namespace,
	}).Info("Exec to pod...")

	err = r.exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *iexec) exec() error {
	// Instantiate loader for kubeconfig file.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"host":    restconfig.Host,
		"apipath": restconfig.APIPath,
	}).Debug("Restconfig")

	if r.container == "" {
		r.container = r.pod.Spec.Containers[0].Name
	}

	req := r.client.CoreV1().RESTClient().
		Post().
		Namespace(r.namespace).
		Resource("pods").
		Name(r.pod.GetName()).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: r.container,
			Command:   r.remoteCmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	log.WithFields(log.Fields{
		"URL": req.URL(),
	}).Debug("Request")

	log.WithFields(log.Fields{
		"container":      r.container,
		"namespace":      r.namespace,
		"remote command": r.remoteCmd,
		"search filter":  r.search,
	}).Debug("iexec struct values...")

	exec, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		return err
	}

	// Put the terminal into raw mode to prevent it echoing characters twice.
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		return err
	}
	defer func() {
		err := terminal.Restore(0, oldState)
		if err != nil {
			log.Error(err)
		}
	}()

	// Connect this process' std{in,out,err} to the remote shell process.
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	})
	if err != nil {
		return err
	}

	fmt.Println()
	return nil
}
