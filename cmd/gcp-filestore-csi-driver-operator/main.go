package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"
	"k8s.io/utils/clock"

	"github.com/openshift/gcp-filestore-csi-driver-operator/pkg/operator"
	"github.com/openshift/gcp-filestore-csi-driver-operator/pkg/version"
)

func main() {
	command := NewOperatorCommand()
	code := cli.Run(command)
	os.Exit(code)
}

func NewOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp-filestore-csi-driver-operator",
		Short: "OpenShift GCP Filestore CSI Driver Operator",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	ctrlCmd := controllercmd.NewControllerCommandConfig(
		"gcp-filestore-csi-driver-operator",
		version.Get(),
		operator.RunOperator,
		clock.RealClock{},
	).NewCommand()
	ctrlCmd.Use = "start"
	ctrlCmd.Short = "Start the GCP Filestore CSI Driver Operator"

	// Inject cluster TLS profile into the operator's own HTTPS endpoint before
	// controllercmd starts the server. This is an OLM-managed operator so it
	// must read the APIServer CR itself — CVO/CSO do not inject TLS config.
	// See: openshift/enhancements#1910
	originalPreRunE := ctrlCmd.PersistentPreRunE
	ctrlCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if originalPreRunE != nil {
			if err := originalPreRunE(cmd, args); err != nil {
				return err
			}
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		configPath, err := operator.WriteOperatorTLSConfig(ctx)
		if err != nil {
			return fmt.Errorf("failed to configure TLS profile: %w", err)
		}
		if configPath != "" {
			if err := cmd.Flags().Set("config", configPath); err != nil {
				return fmt.Errorf("failed to set --config flag: %w", err)
			}
		}
		return nil
	}

	cmd.AddCommand(ctrlCmd)

	return cmd
}
