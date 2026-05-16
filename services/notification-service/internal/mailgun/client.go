package mailgun

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)


type Client struct {
	apiKey     string
	domain     string
	from       string
	httpClient *http.Client
}


func NewClient(apiKey, domain, from string) *Client {
	return &Client{
		apiKey: apiKey,
		domain: domain,
		from:   from,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}


func (c *Client) Send(to, subject, body string) error {
	endpoint := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", c.domain)

	form := url.Values{}
	form.Set("from", c.from)
	form.Set("to", to)
	form.Set("subject", subject)
	form.Set("text", body)

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("mailgun: build request: %w", err)
	}
	req.SetBasicAuth("api", c.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mailgun: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("mailgun: non-2xx response %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}


func (c *Client) IsConfigured() bool {
	return c.apiKey != "" && c.domain != ""
}
