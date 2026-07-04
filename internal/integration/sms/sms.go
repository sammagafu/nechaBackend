package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Config configures a provider-agnostic HTTP SMS gateway. When Enabled is
// false or APIURL is empty, messages are logged only (mirrors the email client).
type Config struct {
	Enabled bool
	APIURL  string
	APIKey  string
	Sender  string
	AdminTo string
}

type Client struct {
	cfg  Config
	http *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: 15 * time.Second}}
}

// Send delivers an SMS to a single recipient. A blank recipient is a no-op so
// callers can pass optional hotel numbers without guarding every call.
func (c *Client) Send(to, body string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return nil
	}
	if !c.cfg.Enabled || c.cfg.APIURL == "" {
		log.Printf("[sms] to=%s body=%q (SMS disabled, logged only)", to, body)
		return nil
	}
	payload, err := json.Marshal(map[string]string{
		"from": c.cfg.Sender,
		"to":   to,
		"text": body,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.cfg.APIURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("sms send failed: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) NotifyAdmin(body string) error {
	return c.Send(c.cfg.AdminTo, body)
}
