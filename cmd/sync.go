package cmd

import (
	"context"
	"fmt"

	"github.com/ccojocar/gh2es/pkg/elastic"
	"github.com/ccojocar/gh2es/pkg/github"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	organisationFlag = "organisation"
	repositoryFlag   = "repository"
	stateFlag        = "state"
	issuesPerPage    = 10
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync the issues from a GitHub repository into an Elasticsearch index",
	Long:  "Sync all issues from a GitHub repository into an Elasticsearch index. The issues are stored as documents.",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := github.ParseToken()
		if err != nil {
			return errors.Wrap(err, "parsing GitHub token from environment")
		}
		if token == "" {
			return fmt.Errorf("no valid GitHub token found (set up a token in GITHUB_TOKEN environment variable or ~/.config/hub)")
		}
		endpoint, err := parseRequiredStringFlag(cmd, endpointFlag)
		if err != nil {
			return err
		}
		index, err := parseRequiredStringFlag(cmd, indexNameFlag)
		if err != nil {
			return err
		}
		state, err := parseStateFlag(cmd)
		if err != nil {
			return err
		}
		organisation, err := parseRequiredStringFlag(cmd, organisationFlag)
		if err != nil {
			return err
		}
		repository, err := parseRequiredStringFlag(cmd, repositoryFlag)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		issueLister := github.IssueLister{
			Organisation: organisation,
			Repository:   repository,
			Token:        token,
		}
		syncer := elastic.Syncer{
			Endpoint: endpoint,
			Index:    index,
		}
		issuesCh, errChList := issueLister.ListStream(ctx, issuesPerPage, state)
		errChSync := syncer.Run(ctx, issuesCh, issuesPerPage)
		for {
			select {
			case err := <-errChList:
				return err
			case err := <-errChSync:
				return err
			}
		}
	},
}

func parseStateFlag(cmd *cobra.Command) (github.IssueState, error) {
	state, err := parseRequiredStringFlag(cmd, stateFlag)
	if err != nil {
		return github.IssueStateAll, err
	}
	switch state {
	case "open":
	case "closed":
	case "all":
		return github.IssueState(state), nil
	}
	return github.IssueStateAll, fmt.Errorf("invalid state %q", state)
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringP(endpointFlag, "e", "", "Elasticsearch endpoint address")
	syncCmd.Flags().StringP(organisationFlag, "o", "", "GitHub organistion where the repository is hosted")
	syncCmd.Flags().StringP(repositoryFlag, "r", "", "Github repository from where the issues are going to be synced up")
	syncCmd.Flags().StringP(indexNameFlag, "i", "", "Elasticsearch index name where the GitHub issues will be syncked")
	syncCmd.Flags().StringP(stateFlag, "s", "all", "Indicates the state of the GitHub issues beeing synced (open|closed|all)")
}
