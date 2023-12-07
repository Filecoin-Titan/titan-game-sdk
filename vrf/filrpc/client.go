package filrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"golang.org/x/xerrors"
)

type request struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// Response defines a JSON RPC response from the spec
// http://www.jsonrpc.org/specification#response_object
type response struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	ID      interface{} `json:"id"`
	Error   *respError  `json:"error,omitempty"`
}

type respError struct {
	Code    errorCode       `json:"code"`
	Message string          `json:"message"`
	Meta    json.RawMessage `json:"meta,omitempty"`
}

type errorCode int
type params []interface{}

type Client struct {
	cfg Config
}

func New(opts ...Option) *Client {
	cfg := DefaultOption()

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Client{
		cfg: cfg,
	}
}

func requestLotus(url string, data request) (*response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rsp response
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		return nil, err
	}

	if rsp.Error != nil {
		return nil, xerrors.New(rsp.Error.Message)
	}

	return &rsp, nil
}

// ChainGetTipSetByHeight lotus ChainGetTipSetByHeight api
func (c *Client) ChainGetTipSetByHeight(height int64) (*TipSet, error) {
	serializedParams := params{
		height, nil,
	}

	req := request{
		Jsonrpc: "2.0",
		Method:  "Filecoin.ChainGetTipSetByHeight",
		Params:  serializedParams,
		ID:      1,
	}

	rsp, err := requestLotus(c.cfg.NodeURL, req)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(rsp.Result)
	if err != nil {
		return nil, err
	}

	var ts TipSet
	err = json.Unmarshal(b, &ts)
	if err != nil {
		return nil, err
	}

	return &ts, nil
}

// ChainHead lotus ChainHead api
func (c *Client) ChainHead() (*TipSet, error) {
	req := request{
		Jsonrpc: "2.0",
		Method:  "Filecoin.ChainHead",
		Params:  nil,
		ID:      1,
	}

	rsp, err := requestLotus(c.cfg.NodeURL, req)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(rsp.Result)
	if err != nil {
		return nil, err
	}

	var ts TipSet
	err = json.Unmarshal(b, &ts)
	if err != nil {
		return nil, err
	}

	return &ts, nil
}
