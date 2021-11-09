package cmd

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/gabeduke/kubectl-iexec/pkg/iexec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	iexecLong = `
IExec is an interactive pod and container selector for 'kubectl exec'

Arg[1] will act as a filter, any pods that match will be returned in a list
that the user can select from.

Arg[2...] are the commands to be executed in the container
`
	iexecExample = `
	# select from all pods in the namespace then run: 'kubectl exec sh'
	%[1]s iexec 

	# select from all pods matching [busybox] then run: 'kubectl exec [pod_name] /bin/sh'
	%[1]s iexec busybox

	# select from all pods matching [busybox] then run: 'kubectl exec [pod_name] cat /etc/hosts'
	%[1]s iexec busybox cat /etc/hosts

	# select from all pods matching [multi_container_pod]
	# then select from all containers in pod matching [second_container]
	# then run: 'kubectl exec [pod_name] [container_name] /bin/sh'
	%[1]s iexec multi_container_pod -c second_container
`
)

// IExecOptions configures the Iexec runner
type IExecOptions struct {
	configFlags *genericclioptions.ConfigFlags
	clientCfg   *rest.Config

	configOverrides clientcmd.ConfigOverrides
	allNamespaces   bool
	remoteCmd       []string
	containerFilter string
	lvl             string
	namespace       string
	naked           bool
	vimMode         bool

	genericclioptions.IOStreams
}

// NewIExecOptions provides an instance of IExecOptions with default values.
func NewIExecOptions(streams genericclioptions.IOStreams) *IExecOptions {
	return &IExecOptions{
		configFlags: genericclioptions.NewConfigFlags(true),

		IOStreams: streams,
	}
}

// NewCmdIExec provides a cobra command wrapping IExecOptions.
func NewCmdIExec(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewIExecOptions(streams)

	cmd := &cobra.Command{
		Use:          "iexec [pod filter] [remote command(s)] [flags]",
		Short:        "Interactive remote shell into a Kubernetes Pod",
		Example:      fmt.Sprintf(iexecExample, "kubectl"),
		Long:         iexecLong,
		SilenceUsage: false,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Run(args); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "A", o.allNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.PersistentFlags().StringVarP(&o.containerFilter, "container", "c", "", "Container to search")
	cmd.PersistentFlags().StringVarP(&o.lvl, "log-level", "l", "", "log level (trace|debug|info|warn|error|fatal|panic)")
	cmd.PersistentFlags().BoolVarP(&o.vimMode, "vim-mode", "v", false, "Vim Mode enabled")
	cmd.PersistentFlags().BoolVarP(&o.naked, "naked", "x", false, "Decolorize output")
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func (o *IExecOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error

	o.clientCfg, err = o.configFlags.ToRESTConfig()
	if err != nil {
		return errors.Wrap(err, "unable to get rest client")
	}

	c := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&o.configOverrides)

	o.namespace, _, err = c.Namespace()
	if err != nil {
		return errors.Wrap(err, "unable to get namespace")
	}

	if *o.configFlags.Namespace != "" {
		o.namespace = *o.configFlags.Namespace
	}

	if o.allNamespaces {
		o.namespace = ""
	}

	switch o.lvl {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.ErrorLevel)
	}

	return nil
}

func (o *IExecOptions) Run(args []string) error {
	podFilter := ""

	if len(args) > 0 {
		podFilter = args[0]
	}

	if len(args) > 1 {
		s := len(args)
		o.remoteCmd = append(o.remoteCmd, args[1:s]...)
	} else {
		o.remoteCmd = []string{"/bin/sh"}
	}

	config := &iexec.Config{
		Namespace:       o.namespace,
		Naked:           o.naked,
		VimMode:         o.vimMode,
		PodFilter:       podFilter,
		ContainerFilter: o.containerFilter,
		RemoteCmd:       o.remoteCmd,
	}

	r := iexec.NewIexec(o.clientCfg, config)

	if err := r.Do(); err != nil {
		log.Fatal(err)
	}

	return nil
}
