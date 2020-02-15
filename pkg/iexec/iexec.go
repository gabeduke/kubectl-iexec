package iexec

import (
	"k8s.io/client-go/kubernetes"

	"github.com/manifoldco/promptui"

	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"

	// auth needed for proxy
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type sizeQueue chan remotecommand.TerminalSize

type Iexecer interface {
	Do() error
}

type Config struct {
	Namespace string
	Naked     bool
	VimMode   bool
}

type Iexec struct {
	client    kubernetes.Interface
	config    *Config
	container string
	pod       v1.Pod
	remoteCmd []string
	search    string
}

func NewIexec(podFilter string, containerFilter string, remoteCmd []string, config *Config) *Iexec {

	iexec := Iexec{
		search:    podFilter,
		container: containerFilter,
		remoteCmd: remoteCmd,
		config:    config,
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	kubeconfig, err := clientConfig.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	iexec.client, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"container":      iexec.container,
		"remote command": iexec.remoteCmd,
		"search filter":  iexec.search,
	}).Debug("iexec struct values...")

	log.WithFields(log.Fields{
		"Vim Mode":  config.VimMode,
		"Naked":     config.Naked,
		"Namespace": config.Namespace,
	}).Debug("iexec config values...")

	return &iexec
}

func (r *Iexec) podPrompt(matchingPods []v1.Pod) error {

	if len(matchingPods) == 1 {
		r.pod = matchingPods[0]
		r.config.Namespace = r.pod.GetNamespace()
		fmt.Printf("Found matching pod (%s) with filter: %s\n", r.pod.Name, r.search)
		return nil
	}

	templates := &promptui.SelectTemplates{
		Active:   fmt.Sprintf("Namespace: {{ .Namespace | blue }} | Pod: %s {{ .Name | cyan }}", promptui.IconSelect),
		Inactive: "Namespace: {{ .Namespace | blue }} | Pod: {{ .Name | magenta }}",
		Selected: fmt.Sprintf("Namespace: {{ .Namespace | blue }} | Pod: %s {{ .Name | cyan }}", promptui.IconGood),
	}

	if r.config.Naked {
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
		IsVimMode: r.config.VimMode,
	}

	i, _, err := podsPrompt.Run()
	if err != nil {
		return err
	}

	r.pod = matchingPods[i]
	r.config.Namespace = r.pod.GetNamespace()
	return nil
}

func (r *Iexec) containerPrompt() error {

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

	if r.config.Naked {
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
		IsVimMode: r.config.VimMode,
	}

	c, _, err := containersPrompt.Run()
	if err != nil {
		return err
	}

	r.container = containers[c]

	return nil
}

func (r *Iexec) Do() error {
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
		"namespace": r.config.Namespace,
	}).Info("Exec to pod...")

	err = r.exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *Iexec) exec() error {
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
		Namespace(r.config.Namespace).
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
		"namespace":      r.config.Namespace,
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

	termWidth, termHeight, _ := terminal.GetSize(0)
	termSize := remotecommand.TerminalSize{Width: uint16(termWidth), Height: uint16(termHeight)}
	s := make(sizeQueue, 1)
	s <- termSize

	defer func() {
		err := terminal.Restore(0, oldState)
		if err != nil {
			log.Error(err)
		}
	}()

	// Connect this process' std{in,out,err} to the remote shell process.
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
		Tty:               true,
		TerminalSizeQueue: s,
	})
	if err != nil {
		return err
	}

	fmt.Println()
	return nil
}

func (s sizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-s
	if !ok {
		return nil
	}
	return &size
}
