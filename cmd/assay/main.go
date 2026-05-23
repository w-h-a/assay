package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/w-h-a/assay/internal/handler/cli"
	"github.com/w-h-a/assay/internal/service"
)

func main() {
	svc := service.New()
	h := cli.New(svc)

	root := &cobra.Command{
		Use:           "assay",
		Short:         "Specification language for software behavior",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print assay version",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return h.Version(cmd.OutOrStdout())
			},
		},
		&cobra.Command{
			Use:   "check <spec.assay>",
			Short: "Validate a spec file",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return h.Check(cmd.ErrOrStderr(), args[0])
			},
		},
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
