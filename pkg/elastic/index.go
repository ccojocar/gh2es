package elastic

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Index keeps various details related to an index
type Index struct {
	// Name index name
	Name string
	// Body contains the index definition
	Body io.Reader
}

// Indexer manages the indices creation in Elasticsearch service
type Indexer struct {
	// Endpoint the address of the Elasticsearch service
	Endpoint string
	// Indices list of indexes which needs to be managed
	Indices []Index
}

// cleanup cleans up any existing index in order to be able to recreate it
func (i *Indexer) cleanup(ctx context.Context) error {
	es, err := client(i.Endpoint)
	if err != nil {
		return err
	}
	var result error
	for _, index := range i.Indices {
		resp, err := es.Indices.Exists([]string{index.Name})
		if err == nil && resp.StatusCode < 400 {
			resp, err := es.Indices.Delete([]string{index.Name}, es.Indices.Delete.WithContext(ctx))
			if err != nil {
				result = multierror.Append(errors.Wrapf(err, "deleting exising index %q", index.Name))
			}
			resp.Body.Close()
		}
		resp.Body.Close()
	}
	return result
}

// Create creates a new index template in Elasticsearch
func (i *Indexer) Create(ctx context.Context) error {
	es, err := client(i.Endpoint)
	if err != nil {
		return err
	}
	if err := i.cleanup(ctx); err != nil {
		return errors.Wrap(err, "cleaning up existing indices")
	}
	var result error
	for _, index := range i.Indices {
		resp, err := es.Indices.Create(index.Name,
			es.Indices.Create.WithBody(index.Body),
			es.Indices.Create.WithContext(ctx),
		)
		if err != nil {
			result = multierror.Append(errors.Wrapf(err, "creating index %q", index.Name))
		}
		if resp.StatusCode >= 400 {
			result = multierror.Append(fmt.Errorf("failed to create %q index: %s", index.Name, parseErrorFromResponse(resp)))
		}
		resp.Body.Close()
	}
	return result
}

//Info retrieves the index info from Elasticsearch service
func (i *Indexer) Info(ctx context.Context) (string, error) {
	es, err := client(i.Endpoint)
	if err != nil {
		return "", err
	}
	var indices []string
	for _, index := range i.Indices {
		indices = append(indices, index.Name)
	}
	resp, err := es.Indices.Get(indices,
		es.Indices.Get.WithContext(ctx),
		es.Indices.Get.WithPretty(),
	)
	if err != nil {
		return "", errors.Wrapf(err, "getting the indicies: %v", indices)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("failed to get %v indecies: %s", indices, parseErrorFromResponse(resp))
	}

	info, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "reading the indices %v details from reponse", indices)
	}

	return string(info), nil
}
