package github

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

const (
	githubTokenEnv     = "GITHUB_TOKEN"
	hubFile            = ".config/hub"
	hubGithubConfigKey = "github.com"
)

// client creates a new GitHub client using the token provided as an argument
func client(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// ParseToken parse a token either from an environment variable or the first github token
// availalbe in the config file of hub tool
func ParseToken() (string, error) {
	token := os.Getenv(githubTokenEnv)
	if token != "" {
		return token, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "getting the current system user")
	}

	hubTokenFile := filepath.Join(u.HomeDir, hubFile)
	data, err := ioutil.ReadFile(hubTokenFile)
	if err != nil {
		return "", errors.Wrapf(err, "searching hub token into %q", hubTokenFile)
	}
	var hubConfigs map[string][]struct {
		OAuthToken string `yaml:"oauth_token"`
		Protocol   string `yaml:"protocol"`
		User       string `yaml:"user"`
	}
	if err := yaml.Unmarshal(data, &hubConfigs); err != nil {
		return "", errors.Wrapf(err, "unmarshaling the hub config from file %q", hubTokenFile)
	}
	githubConfigs, ok := hubConfigs[hubGithubConfigKey]
	if !ok {
		return "", fmt.Errorf("no Github config with key %q found in the hub file %q", hubGithubConfigKey, hubTokenFile)
	}
	if len(githubConfigs) == 0 {
		return "", fmt.Errorf("empty Github config in hub file %q", hubTokenFile)
	}
	return githubConfigs[0].OAuthToken, nil
}
