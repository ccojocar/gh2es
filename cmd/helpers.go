package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func parseRequiredStringFlag(cmd *cobra.Command, flag string) (string, error) {
	value, err := cmd.Flags().GetString(flag)
	if err != nil {
		return "", errors.Wrapf(err, "parsing %q flag", flag)
	}
	if value == "" {
		return "", fmt.Errorf("--%s flag is required", flag)
	}
	return value, nil
}
