package cmd

import (
	"fmt"

	"github.com/ccojocar/gh2es/pkg/elastic"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	endpointFlag = "endpoint"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print some information from Elasticsearch service",
	Long:  "Print some information from Elasticsearch service with contains various details about the cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint, err := parseRequiredStringFlag(cmd, endpointFlag)
		if err != nil {
			return err
		}
		sync := &elastic.Syncer{
			Endpoint: endpoint,
		}
		info, err := sync.Info()
		if err != nil {
			return errors.Wrap(err, "getting the info from Elasticsearch service")
		}
		fmt.Println(info)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringP(endpointFlag, "e", "", "Elasticsearch endpoint address")
}
