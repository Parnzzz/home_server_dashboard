package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type PasswordResetMailer interface {
	Available() bool
	SendCode(email string, code string) error
}

type SMTPPasswordResetMailer struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPPasswordResetMailer(
	host string,
	port int,
	username string,
	password string,
	from string,
) *SMTPPasswordResetMailer {
	return &SMTPPasswordResetMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (m *SMTPPasswordResetMailer) Available() bool {
	return strings.TrimSpace(m.host) != "" && strings.TrimSpace(m.from) != ""
}

func (m *SMTPPasswordResetMailer) SendCode(email string, code string) error {
	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: Password reset code\r\n"+
			"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n"+
			"Your Home Server Dashboard password reset code is %s.\r\n"+
			"It expires in 15 minutes. If you did not request this, you can ignore this email.\r\n",
		m.from,
		email,
		code,
	)

	address := net.JoinHostPort(m.host, strconv.Itoa(m.port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var conn net.Conn
	var err error
	if m.port == 465 {
		conn, err = tls.DialWithDialer(dialer, "tcp", address, &tls.Config{
			ServerName: m.host,
			MinVersion: tls.VersionTLS12,
		})
	} else {
		conn, err = dialer.Dial("tcp", address)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return err
	}
	defer client.Close()

	if m.port != 465 {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{
			ServerName: m.host,
			MinVersion: tls.VersionTLS12,
		}); err != nil {
			return err
		}
	}

	if m.username != "" {
		if err := client.Auth(smtp.PlainAuth("", m.username, m.password, m.host)); err != nil {
			return err
		}
	}
	if err := client.Mail(m.from); err != nil {
		return err
	}
	if err := client.Rcpt(email); err != nil {
		return err
	}

	messageWriter, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := messageWriter.Write([]byte(message)); err != nil {
		_ = messageWriter.Close()
		return err
	}
	if err := messageWriter.Close(); err != nil {
		return err
	}
	return client.Quit()
}

type ConsolePasswordResetMailer struct {
	logger *log.Logger
}

func NewConsolePasswordResetMailer(logger *log.Logger) *ConsolePasswordResetMailer {
	return &ConsolePasswordResetMailer{logger: logger}
}

func (m *ConsolePasswordResetMailer) Available() bool {
	return m.logger != nil
}

func (m *ConsolePasswordResetMailer) SendCode(email string, code string) error {
	if m.logger == nil {
		return fmt.Errorf("development password reset logger is unavailable")
	}
	m.logger.Printf("[development only] password reset email to %s; code: %s", email, code)
	return nil
}

type UnavailablePasswordResetMailer struct{}

func (UnavailablePasswordResetMailer) Available() bool { return false }

func (UnavailablePasswordResetMailer) SendCode(string, string) error {
	return fmt.Errorf("password reset email is not configured")
}
