package cmd

import (
	"github.com/gabeduke/kubectl-iexec/pkg/cmd"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	flags := pflag.NewFlagSet("kubectl-iexec", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewCmdIExec(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
