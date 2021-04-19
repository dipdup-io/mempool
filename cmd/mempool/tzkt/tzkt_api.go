package tzkt

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// TzKTAPI -
type TzKTAPI struct {
	url    string
	client *http.Client
}

// NewTzKTAPI -
func NewTzKTAPI(baseURL string) *TzKTAPI {
	return &TzKTAPI{
		url:    baseURL,
		client: http.DefaultClient,
	}
}

func (tzkt *TzKTAPI) get(endpoint string, args map[string]string, output interface{}) error {
	u, err := url.Parse(tzkt.url)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, endpoint)

	values := u.Query()
	for key, value := range args {
		values.Add(key, value)
	}
	u.RawQuery = values.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := tzkt.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return json.NewDecoder(resp.Body).Decode(output)
	}

	return errors.New(resp.Status)
}

// GetHead -
func (tzkt *TzKTAPI) GetHead() (head Head, err error) {
	err = tzkt.get("/v1/head", nil, &head)
	return
}

// GetBlock -
func (tzkt *TzKTAPI) GetBlock(level uint64) (b Block, err error) {
	err = tzkt.get(fmt.Sprintf("/v1/blocks/%d", level), nil, &b)
	return
}

// GetBlocks -
func (tzkt *TzKTAPI) GetBlocks(filters map[string]string) (b []Block, err error) {
	err = tzkt.get("/v1/blocks", filters, &b)
	return
}

// GetEndorsements -
func (tzkt *TzKTAPI) GetEndorsements(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/endorsements", filters, &operations)
	return
}

// GetBallots -
func (tzkt *TzKTAPI) GetBallots(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/ballots", filters, &operations)
	return
}

// GetProposals -
func (tzkt *TzKTAPI) GetProposals(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/proposals", filters, &operations)
	return
}

// GetActivations -
func (tzkt *TzKTAPI) GetActivations(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/activations", filters, &operations)
	return
}

// GetDoubleBakings -
func (tzkt *TzKTAPI) GetDoubleBakings(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/double_baking", filters, &operations)
	return
}

// GetDoubleEndorsings -
func (tzkt *TzKTAPI) GetDoubleEndorsings(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/double_endorsing", filters, &operations)
	return
}

// GetNonceRevelations -
func (tzkt *TzKTAPI) GetNonceRevelations(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/nonce_revelations", filters, &operations)
	return
}

// GetDelegations -
func (tzkt *TzKTAPI) GetDelegations(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/delegations", filters, &operations)
	return
}

// GetOriginations -
func (tzkt *TzKTAPI) GetOriginations(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/originations", filters, &operations)
	return
}

// GetTransactions -
func (tzkt *TzKTAPI) GetTransactions(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/transactions", filters, &operations)
	return
}

// GetReveals -
func (tzkt *TzKTAPI) GetReveals(filters map[string]string) (operations []Operation, err error) {
	err = tzkt.get("/v1/operations/reveals", filters, &operations)
	return
}
