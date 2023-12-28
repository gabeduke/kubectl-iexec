package iexec

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	// auth needed for proxy.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type sizeQueue chan remotecommand.TerminalSize

type Iexecer interface {
	Do() error
}

type Config struct {
	Namespace       string
	LabelSelector   string
	Naked           bool
	VimMode         bool
	PodFilter       string
	ContainerFilter string
	RemoteCmd       []string
}

type Iexec struct {
	restConfig *rest.Config
	config     *Config
}

func NewIexec(restConfig *rest.Config, config *Config) *Iexec {
	log.WithFields(log.Fields{
		"containerFilter": config.ContainerFilter,
		"remote command":  config.RemoteCmd,
		"podFilter":       config.PodFilter,
		"Vim Mode":        config.VimMode,
		"Naked":           config.Naked,
		"Namespace":       config.Namespace,
		"LabelSelector":   config.LabelSelector,
	}).Debug("iexec config values...")

	return &Iexec{restConfig: restConfig, config: config}
}

func selectPod(pods []corev1.Pod, config Config) (corev1.Pod, error) {
	if len(pods) == 1 {
		return pods[0], nil
	}

	templates := podTemplate

	if config.Naked {
		templates = podTemplateNaked
	}

	podsPrompt := promptui.Select{
		Label:     "Select Pod",
		Items:     pods,
		Templates: templates,
		IsVimMode: config.VimMode,
	}

	i, _, err := podsPrompt.Run()
	if err != nil {
		return pods[i], errors.Wrap(err, "unable to run prompt")
	}

	return pods[i], nil
}

func containerPrompt(containers []corev1.Container, config Config) (corev1.Container, error) {
	if len(containers) == 1 {
		return containers[0], nil
	}

	templates := containerTemplates

	if config.Naked {
		templates = containerTemplatesNaked
	}

	containersPrompt := promptui.Select{
		Label:     "Select Container",
		Items:     containers,
		Templates: templates,
		IsVimMode: config.VimMode,
	}

	i, _, err := containersPrompt.Run()
	if err != nil {
		return containers[i], errors.Wrap(err, "unable to get prompt")
	}

	return containers[i], nil
}

func (r *Iexec) Do() error {
	client, err := kubernetes.NewForConfig(r.restConfig)
	if err != nil {
		return errors.Wrap(err, "unable to get kubernetes for config")
	}

	pods, err := getAllPods(client, r.config.Namespace, r.config.LabelSelector)
	if err != nil {
		return err
	}

	filteredPods, err := r.matchPods(pods)
	if err != nil {
		return err
	}

	pod, err := selectPod(filteredPods.Items, *r.config)
	if err != nil {
		return err
	}

	containers, err := matchContainers(pod, *r.config)
	if err != nil {
		return err
	}

	container, err := containerPrompt(containers, *r.config)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"pod":       pod.GetName(),
		"container": container.Name,
		"namespace": r.config.Namespace,
	}).Info("Exec to pod...")

	err = exec(r.restConfig, pod, container, r.config.RemoteCmd)
	if err != nil {
		return err
	}
	return nil
}

func exec(restCfg *rest.Config, pod corev1.Pod, container corev1.Container, cmd []string) error {
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return errors.Wrap(err, "unable to get kubernetes client config")
	}

	req := client.CoreV1().RESTClient().
		Post().
		Namespace(pod.GetNamespace()).
		Resource("pods").
		Name(pod.GetName()).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container.Name,
			Command:   cmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	log.WithFields(log.Fields{
		"URL": req.URL(),
	}).Debug("Request")

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return errors.Wrap(err, "unable to execute remote command")
	}

	fd := int(os.Stdin.Fd())

	// Put the terminal into raw mode to prevent it echoing characters twice.
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return errors.Wrap(err, "unable to init terminal")
	}

	termWidth, termHeight, _ := term.GetSize(fd)
	termSize := remotecommand.TerminalSize{Width: uint16(termWidth), Height: uint16(termHeight)}
	s := make(sizeQueue, 1)
	s <- termSize

	defer func() {
		err := term.Restore(fd, oldState)
		if err != nil {
			log.Error(err)
		}
	}()

	// Connect this process' std{in,out,err} to the remote shell process.
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
		Tty:               true,
		TerminalSizeQueue: s,
	})
	if err != nil {
		return errors.Wrap(err, "unable to stream shell process")
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
