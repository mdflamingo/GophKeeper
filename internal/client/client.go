package client

import (
	"github.com/go-resty/resty/v2"
)

type Client struct {
	Client *resty.Client
	Token  string
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
