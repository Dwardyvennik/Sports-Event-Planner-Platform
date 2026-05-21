package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"mime"
	"net"
	netsmtp "net/smtp"
	"regexp"
	"strings"
	"time"
)

const connectTimeout = 10 * time.Second

var htmlTagPattern = regexp.MustCompile(`(?i)<[a-z][a-z0-9]*(\s|>|/)`)

type SMTPClient struct {
	host   string
	port   string
	from   string
	useTLS bool
}

func NewSMTPClient(host, port, from string, useTLS bool) *SMTPClient {
	return &SMTPClient{
		host:   host,
		port:   port,
		from:   from,
		useTLS: useTLS,
	}
}

func (c *SMTPClient) Send(to, subject, body string) error {
	if !c.IsConfigured() {
		return fmt.Errorf("smtp: host is not configured")
	}
	if strings.TrimSpace(c.from) == "" {
		return fmt.Errorf("smtp: from address is not configured")
	}

	port := c.port
	if port == "" {
		port = "1025"
	}

	slog.Default().Info("smtp send",
		"host", c.host,
		"port", port,
		"to", to,
		"subject", subject,
	)

	address := net.JoinHostPort(c.host, port)
	dialer := &net.Dialer{Timeout: connectTimeout}
	conn, err := dialer.DialContext(context.Background(), "tcp", address)
	if err != nil {
		return fmt.Errorf("smtp: connect %s: %w", address, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(connectTimeout)); err != nil {
		return fmt.Errorf("smtp: set deadline: %w", err)
	}

	client, err := netsmtp.NewClient(conn, c.host)
	if err != nil {
		return fmt.Errorf("smtp: create client: %w", err)
	}
	defer client.Close()

	if c.useTLS {
		tlsConfig := &tls.Config{
			ServerName: c.host,
			MinVersion: tls.VersionTLS12,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp: start tls: %w", err)
		}
	}

	if err := client.Mail(c.from); err != nil {
		return fmt.Errorf("smtp: mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp: rcpt to: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp: data: %w", err)
	}

	if _, err := writer.Write(buildMessage(c.from, to, subject, body)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("smtp: write message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp: close data: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp: quit: %w", err)
	}
	return nil
}

func (c *SMTPClient) IsConfigured() bool {
	return strings.TrimSpace(c.host) != ""
}

func buildMessage(from, to, subject, body string) []byte {
	contentType := "text/plain; charset=utf-8"
	messageBody := body
	if looksHTML(body) {
		contentType = "text/html; charset=utf-8"
	}

	var msg bytes.Buffer
	msg.WriteString("From: " + sanitizeHeader(from) + "\r\n")
	msg.WriteString("To: " + sanitizeHeader(to) + "\r\n")
	msg.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", sanitizeHeader(subject)) + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: " + contentType + "\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(messageBody)
	return msg.Bytes()
}

func looksHTML(body string) bool {
	return htmlTagPattern.MatchString(body)
}

func sanitizeHeader(value string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(value)
}
