package iexec

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

// Iexecer remains for backward compatibility if needed
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

// Iexec is our main struct, now with interfaces for UI, K8s fetching, and exec
type Iexec struct {
	restConfig *rest.Config
	config     *Config

	ui        TerminalUI
	k8sClient K8sClient
	exec      ExecRunner
}

// NewIexec constructs the Iexec with your chosen implementations
func NewIexec(restConfig *rest.Config, config *Config, ui TerminalUI, k8s K8sClient, exec ExecRunner) *Iexec {
	log.WithFields(log.Fields{
		"containerFilter": config.ContainerFilter,
		"remote command":  config.RemoteCmd,
		"podFilter":       config.PodFilter,
		"Vim Mode":        config.VimMode,
		"Naked":           config.Naked,
		"Namespace":       config.Namespace,
		"LabelSelector":   config.LabelSelector,
	}).Debug("iexec config values...")

	return &Iexec{
		restConfig: restConfig,
		config:     config,
		ui:         ui,
		k8sClient:  k8s,
		exec:       exec,
	}
}

// Do is our main plugin flow, now purely orchestrating the calls
func (r *Iexec) Do() error {
	// 1. Fetch pods
	podsList, err := r.k8sClient.FetchAllPods(r.config.Namespace, r.config.LabelSelector)
	if err != nil {
		return errors.Wrap(err, "fetchAllPods failed")
	}

	// 2. Match/Filter pods
	filtered, err := r.k8sClient.MatchPods(podsList, *r.config)
	if err != nil {
		return errors.Wrap(err, "matchPods failed")
	}
	pods := filtered.Items
	if len(pods) == 0 {
		return errors.New("no matching pods found")
	}

	// 3. Prompt user to select a pod
	pod, err := r.ui.SelectPod(pods, *r.config)
	if err != nil {
		return errors.Wrap(err, "SelectPod failed")
	}

	// 4. Match containers
	containers, err := r.k8sClient.MatchContainers(pod, *r.config)
	if err != nil {
		return errors.Wrap(err, "matchContainers failed")
	}
	if len(containers) == 0 {
		return errors.New("no matching containers found")
	}

	// 5. Prompt user to select a container
	container, err := r.ui.SelectContainer(containers, *r.config)
	if err != nil {
		return errors.Wrap(err, "SelectContainer failed")
	}

	log.WithFields(log.Fields{
		"pod":       pod.GetName(),
		"container": container.Name,
		"namespace": r.config.Namespace,
	}).Info("Exec to pod...")

	// 6. Execute remote command
	if err := r.exec.RunExec(r.restConfig, pod, container, r.config.RemoteCmd); err != nil {
		return errors.Wrap(err, "RunExec failed")
	}

	return nil
}
