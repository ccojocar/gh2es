package elastic

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/ccojocar/gh2es/pkg/github"
	"github.com/pkg/errors"
)

// Sync the documents from a docs source into Elasticsearch
type Syncer struct {
	// Endpoint address of the Elsticserach instance to use for syncing the docs
	Endpoint string
	// Index the name of the index where to sync the GitHub issues
	Index string
}

// Info returns some info from Elasticsearch instance
func (s *Syncer) Info() (string, error) {
	es, err := client(s.Endpoint)
	if err != nil {
		return "", err
	}
	info, err := es.Info()
	if err != nil {
		return "", errors.Wrap(err, "getting elastic info")
	}
	return info.String(), nil
}

// updateIssueToIndex updates an issues into Elasticsearch index using the Issue ID
// as the document ID. Also the document type is set to 'issue'.
func (s *Syncer) updateIssueToIndex(ctx context.Context, issue *github.Issue) error {
	es, err := client(s.Endpoint)
	if err != nil {
		return err
	}
	body, err := issue.ToReader()
	if err != nil {
		return err
	}
	resp, err := es.Index(s.Index,
		body,
		es.Index.WithContext(ctx),
		es.Index.WithDocumentID(strconv.Itoa(issue.Number)),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return parseErrorFromResponse(resp)
	}

	fmt.Printf("syncked: %s/%s/issues/%d\n", issue.Organisation, issue.Repository, issue.Number)

	return nil
}

// runWorker runs a worker which uploads the GitHub issue into Elasticserach index
func (s *Syncer) runWorker(ctx context.Context, wg *sync.WaitGroup, chIssue chan *github.Issue, errCh chan error) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			break
		case issue, ok := <-chIssue:
			if !ok {
				return
			}
			err := s.updateIssueToIndex(ctx, issue)
			if err != nil {
				errCh <- err
			}
		}
	}
}

// Run starts the syncker which will spawn multiple workers which will effective upload the GitHub issues
// into Elasticserch index
func (s *Syncer) Run(ctx context.Context, issuesCh chan []*github.Issue, workers int) chan error {
	errCh := make(chan error)
	go func() {
		issueCh := make(chan *github.Issue, workers)
		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				s.runWorker(ctx, &wg, issueCh, errCh)
			}()
		}
		defer close(issueCh)
		defer wg.Wait()
		for {
			select {
			case <-ctx.Done():
				break
			case issues, ok := <-issuesCh:
				if !ok {
					return
				}
				for _, issue := range issues {
					issueCh <- issue
				}
			}
		}
	}()
	return errCh
}
