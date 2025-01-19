package iexec

import (
	"os"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// PromptUITerminal is the real interactive UI using promptui
type PromptUITerminal struct {
	// Add fields if needed (e.g., templates)
}

// NewPromptUITerminal constructor
func NewPromptUITerminal() *PromptUITerminal {
	return &PromptUITerminal{}
}

func (p *PromptUITerminal) SelectPod(pods []corev1.Pod, config Config) (corev1.Pod, error) {
	// Single item => return immediately
	if len(pods) == 1 {
		return pods[0], nil
	}

	// Multiple items => open /dev/tty
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return corev1.Pod{}, errors.Wrap(err, "failed to open /dev/tty for pod selection")
	}
	defer tty.Close()

	var templates *promptui.SelectTemplates
	if config.Naked {
		templates = podTemplateNaked
	} else {
		templates = podTemplate
	}

	prompt := promptui.Select{
		Label:     "Select Pod",
		Items:     pods,
		Stdout:    tty, // Render to the TTY
		IsVimMode: config.VimMode,
		Templates: templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return corev1.Pod{}, errors.Wrap(err, "promptui pod selection failed")
	}
	return pods[i], nil
}

func (p *PromptUITerminal) SelectContainer(containers []corev1.Container, config Config) (corev1.Container, error) {
	if len(containers) == 1 {
		return containers[0], nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return corev1.Container{}, errors.Wrap(err, "failed to open /dev/tty for container selection")
	}
	defer tty.Close()

	var templates *promptui.SelectTemplates
	if config.Naked {
		templates = containerTemplatesNaked
	} else {
		templates = containerTemplates
	}

	prompt := promptui.Select{
		Label:     "Select Container",
		Items:     containers,
		Stdout:    tty,
		IsVimMode: config.VimMode,
		Templates: templates,
	}
	i, _, err := prompt.Run()
	if err != nil {
		return corev1.Container{}, errors.Wrap(err, "promptui container selection failed")
	}
	return containers[i], nil
}
