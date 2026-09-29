package email

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type Config struct {
	Enabled  bool
	Host     string
	Port     string
	Username string
	Password string
	From     string
	AdminTo  string
}

type Client struct {
	cfg Config
}

func NewClient(cfg Config) *Client {
	if cfg.From == "" {
		cfg.From = "noreply@necha.africa"
	}
	if cfg.AdminTo == "" {
		cfg.AdminTo = "info@necha.africa"
	}
	return &Client{cfg: cfg}
}

func (c *Client) Send(to, subject, body string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return nil
	}
	if !c.cfg.Enabled || c.cfg.Host == "" {
		log.Printf("[email] to=%s subject=%q (SMTP disabled, logged only)", to, subject)
		return nil
	}
	addr := fmt.Sprintf("%s:%s", c.cfg.Host, c.cfg.Port)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", c.cfg.From, to, subject, body))
	auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)
	return smtp.SendMail(addr, auth, c.cfg.From, []string{to}, msg)
}

func (c *Client) NotifyAdmin(subject, body string) error {
	return c.Send(c.cfg.AdminTo, subject, body)
}
