package elastic

import (
	"fmt"
	"io/ioutil"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/pkg/errors"
)

// client creats a new Elasticsearch client for given endpoint
func client(endpoint string) (*elasticsearch.Client, error) {
	config := elasticsearch.Config{
		Addresses: []string{endpoint},
	}
	es, err := elasticsearch.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating elastic client")
	}
	return es, nil
}

// parseErrorFromResponse parse the error message from the Elasticsearch response
func parseErrorFromResponse(resp *esapi.Response) error {
	var errMsg string
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		errMsg = fmt.Sprintf("error=%s", string(body))
	}
	return fmt.Errorf("status code=%d %s", resp.StatusCode, errMsg)
}
