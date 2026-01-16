package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Config holds mail server configuration
type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	UseTLS   bool   `mapstructure:"use_tls"`
	MockSend bool   `mapstructure:"mock_send"`
}

// Client is the mail client
type Client struct {
	config *Config
}

// NewClient creates a new mail client
func NewClient(cfg *Config) *Client {
	if cfg.From == "" {
		cfg.From = cfg.User
	}
	if cfg.Port == 0 {
		cfg.Port = 587 // Default STARTTLS port
	}
	return &Client{config: cfg}
}

// SendMail sends an email
func (c *Client) SendMail(to []string, subject, body string) error {
	return c.SendHTMLMail(to, subject, body, false)
}

// SendHTMLMail sends an HTML email
func (c *Client) SendHTMLMail(to []string, subject, body string, isHTML bool) error {
	// Mock mode - skip actual sending
	if c.config.MockSend {
		return nil
	}

	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: %s; charset=UTF-8\r\n"+
		"\r\n%s",
		c.config.From,
		strings.Join(to, ","),
		subject,
		contentType,
		body,
	)

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// Port 465 uses implicit TLS (SSL)
	if c.config.Port == 465 {
		return c.sendMailSSL(addr, to, []byte(msg))
	}

	// Port 587 or 25 uses STARTTLS
	return c.sendMailSTARTTLS(addr, to, []byte(msg))
}

// sendMailSSL sends mail using implicit TLS (port 465)
func (c *Client) sendMailSSL(addr string, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName:         c.config.Host,
		InsecureSkipVerify: false,
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("create client failed: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", c.config.User, c.config.Password, c.config.Host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	if err = client.Mail(c.config.From); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("rcpt to failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %w", err)
	}

	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	w.Close()
	client.Quit()
	return nil
}

// sendMailSTARTTLS sends mail using STARTTLS (port 587/25)
func (c *Client) sendMailSTARTTLS(addr string, to []string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("create client failed: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName:         c.config.Host,
		InsecureSkipVerify: false,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	auth := smtp.PlainAuth("", c.config.User, c.config.Password, c.config.Host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	if err = client.Mail(c.config.From); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("rcpt to failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %w", err)
	}

	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	w.Close()
	client.Quit()
	return nil
}
