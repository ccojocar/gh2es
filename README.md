# gh2es

This is a tool which allows to upload the issues from a GitHub repository into an Elasticsearch index.

## Installation

You can install this tool with:

```sh
go get -u https://github.com/ccojocar/gh2es
```

## Usage

The Elasticsearch index needs to be initialised first. It is recommended to keep one index per GitHub repository. You can achieve this by running the following command:

 ```sh
  gh2es init --endpoint https://<elasticserach-addresss> --file=index.json --index=github

 ```

This will create an index which has the name `github` with the properties described in the [index.json](index.json) file.

The issues can now be synced up from a GitHub repository into the index which has just been created. You can do this by running:

```sh
gh2es sync --endpoint https://<elasticsearch-address> --index=github --organisation=<github-org-name> --repository=<github-repository>
```

## GitHub Authentication

The tools will parse the Github token either from the `GITHUB_TOKEN` environment variable, or from the configuration of the [github/hub](https://github.com/github/hub) tool. This configuration is usually stored into `~/.config/hub` file.
