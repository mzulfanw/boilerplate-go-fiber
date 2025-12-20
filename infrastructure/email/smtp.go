package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	emailservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/email"
)

type SMTPClient struct {
	host               string
	port               int
	username           string
	password           string
	from               string
	useTLS             bool
	useStartTLS        bool
	insecureSkipVerify bool
	timeout            time.Duration
}

func NewSMTP(cfg config.Config) (*SMTPClient, error) {
	if !cfg.EmailEnabled {
		return nil, nil
	}
	if strings.TrimSpace(cfg.SMTPHost) == "" {
		return nil, errors.New("email: SMTP_HOST is required")
	}
	if cfg.SMTPPort == 0 {
		return nil, errors.New("email: SMTP_PORT is required")
	}
	if strings.TrimSpace(cfg.EmailFrom) == "" {
		return nil, errors.New("email: EMAIL_FROM is required")
	}

	timeout := cfg.SMTPTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &SMTPClient{
		host:               cfg.SMTPHost,
		port:               cfg.SMTPPort,
		username:           cfg.SMTPUsername,
		password:           cfg.SMTPPassword,
		from:               cfg.EmailFrom,
		useTLS:             cfg.SMTPTLSEnabled,
		useStartTLS:        cfg.SMTPStartTLSEnabled,
		insecureSkipVerify: cfg.SMTPInsecureSkipVerify,
		timeout:            timeout,
	}, nil
}

func (s *SMTPClient) Send(ctx context.Context, msg emailservice.Message) error {
	if s == nil {
		return errors.New("email: SMTP client is nil")
	}
	if len(msg.To) == 0 {
		return errors.New("email: recipient is empty")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	dialer := &net.Dialer{Timeout: s.timeout}

	var conn net.Conn
	var err error
	if s.useTLS {
		tlsConfig := &tls.Config{
			ServerName:         s.host,
			InsecureSkipVerify: s.insecureSkipVerify,
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("email: dial SMTP: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("email: create SMTP client: %w", err)
	}
	defer client.Close()

	if !s.useTLS && s.useStartTLS {
		tlsConfig := &tls.Config{
			ServerName:         s.host,
			InsecureSkipVerify: s.insecureSkipVerify,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("email: starttls failed: %w", err)
		}
	}

	if s.username != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("email: auth failed: %w", err)
		}
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("email: MAIL FROM failed: %w", err)
	}
	for _, recipient := range msg.To {
		if err := client.Rcpt(strings.TrimSpace(recipient)); err != nil {
			return fmt.Errorf("email: RCPT TO failed: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("email: DATA failed: %w", err)
	}

	if err := writeMessage(writer, s.from, msg); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("email: closing DATA failed: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("email: quit failed: %w", err)
	}

	return nil
}

func writeMessage(w io.Writer, from string, msg emailservice.Message) error {
	headers := map[string]string{
		"From":         from,
		"To":           strings.Join(msg.To, ", "),
		"Subject":      msg.Subject,
		"MIME-Version": "1.0",
		"Content-Type": msg.ContentType,
	}

	for key, value := range msg.Headers {
		if strings.TrimSpace(key) == "" {
			continue
		}
		headers[key] = value
	}

	builder := &strings.Builder{}
	for key, value := range headers {
		builder.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	builder.WriteString("\r\n")
	builder.WriteString(msg.Body)

	if _, err := fmt.Fprint(w, builder.String()); err != nil {
		return fmt.Errorf("email: write body failed: %w", err)
	}
	return nil
}
