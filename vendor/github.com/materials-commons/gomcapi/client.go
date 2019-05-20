package mcapi

import (
	"crypto/tls"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/materials-commons/gomcapi/pkg/urlpath"
	"gopkg.in/resty.v1"
)

type Client struct {
	APIKey  string
	BaseURL string
}

var ErrAuth = errors.New("authentication")

var tlsConfig = tls.Config{InsecureSkipVerify: true}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: urlpath.Join(baseURL, "v3"),
	}
}

func (c *Client) r() *resty.Request {
	return resty.SetTLSClientConfig(&tlsConfig).R().SetQueryParam("apikey", c.APIKey)
}

func (c *Client) join(paths ...string) string {
	return urlpath.Join(c.BaseURL, paths...)
}

func (c *Client) post(result, body interface{}, paths ...string) error {
	p := c.join(paths...)
	resp, err := c.r().SetResult(&result).SetBody(body).Post(p)
	return c.getAPIError(p, resp, err)
}

func (c *Client) getAPIError(p string, resp *resty.Response, err error) error {
	switch {
	case err != nil:
		return err
	case resp.RawResponse.StatusCode == 401:
		return ErrAuth
	case resp.RawResponse.StatusCode > 299:
		return c.toErrorFromResponse(p, resp)
	default:
		return nil
	}
}

func (c *Client) toErrorFromResponse(p string, resp *resty.Response) error {
	var er struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(resp.Body(), &er); err != nil {
		return errors.New(fmt.Sprintf("mcapi '%s' (HTTP Status: %d)- unable to parse json error response: %s", p, resp.RawResponse.StatusCode, err))
	}

	return errors.New(fmt.Sprintf("mcapi '%s' (HTTP Status: %d)- %s", p, resp.RawResponse.StatusCode, er.Error))
}
