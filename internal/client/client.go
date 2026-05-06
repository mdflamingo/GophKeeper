package client

import (
	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

type RequestBody struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewClient(baseURL string) *Client {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetHeader("Content-Type", "application/json")

	return &Client{
		client: client,
	}
}
