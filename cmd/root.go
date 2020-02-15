package cmd

import (
	"fmt"
	"github.com/gabeduke/kubectl-iexec/pkg/iexec"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	remoteCmd []string
	containerFilter string
	lvl       string
	naked     bool
	namespace string
	vimMode   bool
)

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

		config := &iexec.Config{
			Namespace: namespace,
			Naked: naked,
			VimMode: vimMode,
		}

		podFilter := args[0]

		if len(args) > 1 {
			s := len(args)
			remoteCmd = append(remoteCmd, args[1:s]...)
		} else {
			remoteCmd = []string{"/bin/sh"}
		}

		r := iexec.NewIexec(podFilter, containerFilter, remoteCmd, config)

		if err := r.Do(); err != nil {
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
	rootCmd.PersistentFlags().StringVarP(&containerFilter, "container", "c", "", "Container to search")
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
