package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	container string
	lvl       string
	naked     bool
	namespace string
	vimMode   bool
)

type iexec struct {
	client    kubernetes.Interface
	container string
	namespace string
	pod       v1.Pod
	remoteCmd []string
	search    string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iexec [pod filter] [remote command(s)]",
	Short: "kubectl-iexec is CLI for remote shell into a Kubernetes Pod",
	Args:  cobra.MinimumNArgs(1),
	Long: `
Kubectl-iexec is an interactive pod and container selector for 'kubectl exec'

Arg[1] will act as a filter, any pods that match will be returned in a list
that the user can select from. Subsequent args make up the array of commands that
should be executed on the pod.

example:
kubectl iexec busybox /bin/sh

example:
kubectl iexec busybox cat /etc/hosts
`,
	Run: func(cmd *cobra.Command, args []string) {
		r := iexec{}

		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)

		kubeconfig, err := clientConfig.ClientConfig()
		if err != nil {
			log.Fatal(err)
		}

		r.client, err = kubernetes.NewForConfig(kubeconfig)
		if err != nil {
			log.Fatal(err)
		}

		r.search = args[0]

		if len(args) > 1 {
			s := len(args)
			r.remoteCmd = append(r.remoteCmd, args[1:s]...)
		} else {
			r.remoteCmd = []string{"/bin/sh"}
		}

		if container != "" {
			r.container = container
		}

		if namespace != "" {
			r.namespace = namespace
		}

		log.WithFields(log.Fields{
			"container":      r.container,
			"namespace":      r.namespace,
			"remote command": r.remoteCmd,
			"search filter":  r.search,
		}).Debug("iexec struct values...")

		if err := r.do(); err != nil {
			log.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger)

	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace to search")
	rootCmd.PersistentFlags().StringVarP(&container, "container", "c", "", "Container to search")
	rootCmd.PersistentFlags().StringVarP(&lvl, "log-level", "l", "", "log level (trace|debug|info|warn|error|fatal|panic)")
	rootCmd.PersistentFlags().BoolVarP(&vimMode, "vim-mode", "v", false, "Vim Mode enabled")
	rootCmd.PersistentFlags().BoolVarP(&naked, "naked", "x", false, "Decolorize output")
}

func initLogger() {
	switch lvl {
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
}
