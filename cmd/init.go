package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ccojocar/gh2es/pkg/elastic"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	indexNameFlag = "index"
	indexFile     = "file"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates an index in the Elasticsearch service",
	Long:  "Creates an index in the Elsticsearch service",
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint, err := parseRequiredStringFlag(cmd, endpointFlag)
		if err != nil {
			return err
		}
		indexName, err := parseRequiredStringFlag(cmd, indexNameFlag)
		if err != nil {
			return err
		}
		indexFile, err := parseRequiredStringFlag(cmd, indexFile)
		if err != nil {
			return err
		}

		fileInfo, err := os.Stat(indexFile)
		if os.IsNotExist(err) {
			return fmt.Errorf("provided index file %q does not exist", indexFile)
		}
		if fileInfo.IsDir() {
			return fmt.Errorf("provided index file %q is a directoy", indexFile)
		}

		file, err := os.Open(indexFile)
		if err != nil {
			return errors.Wrapf(err, "opening the index file %q", indexFile)
		}
		indexer := elastic.Indexer{
			Endpoint: endpoint,
			Indices: []elastic.Index{
				elastic.Index{
					Name: indexName,
					Body: file,
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := indexer.Create(ctx); err != nil {
			return errors.Wrapf(err, "creating index %q", indexName)
		}

		fmt.Printf("Index %q created\n", indexName)

		info, err := indexer.Info(ctx)
		if err != nil {
			return errors.Wrap(err, "getting the index details from Elaticsearch")
		}
		fmt.Println("Details:")
		fmt.Printf("%s\n", info)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP(endpointFlag, "e", "", "Elasticsearch endpoint address")
	initCmd.Flags().StringP(indexNameFlag, "n", "", "Elasticsearch index name")
	initCmd.Flags().StringP(indexFile, "f", "", "JSON file containing the Elasticsearch index template")
}
