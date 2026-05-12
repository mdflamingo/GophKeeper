package client

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	Client      *resty.Client
	Token       string
	MasterPass  string
	PassExpires time.Time
}

func NewClient(baseURL string) *Client {
	client := resty.New()
	client.SetBaseURL(baseURL)

	return &Client{
		Client: client,
	}
}

func (c *Client) SetToken(token string) {
	c.Token = token
	if token != "" {
		c.Client.SetAuthToken(token)
	}
}

func (c *Client) GetToken() string {
	return c.Token
}

func (c *Client) GetCachedMasterPassword() (string, bool) {
	if c.MasterPass != "" && time.Now().Before(c.PassExpires) {
		return c.MasterPass, true
	}
	return "", false
}

func (c *Client) SetMasterPassword(pass string) {
	c.MasterPass = pass
	c.PassExpires = time.Now().Add(30 * time.Minute)
}

func (c *Client) ClearMasterPassword() {
	c.MasterPass = ""
	c.PassExpires = time.Time{}
}
