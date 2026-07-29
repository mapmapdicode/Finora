package email

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"wealthos-backend/internal/config"
)

// VerificationSender delivers the one-time code used to activate an account.
type VerificationSender interface {
	SendVerificationCode(ctx context.Context, recipient, code string) error
}

type smtpVerificationSender struct {
	host, port, username, password, from, fromName string
}

type developmentVerificationSender struct{}

func NewVerificationSender(cfg *config.Config) VerificationSender {
	if cfg != nil && strings.TrimSpace(cfg.SMTPHost) != "" && strings.TrimSpace(cfg.SMTPFrom) != "" {
		return &smtpVerificationSender{
			host: cfg.SMTPHost, port: cfg.SMTPPort, username: cfg.SMTPUsername,
			password: cfg.SMTPPassword, from: cfg.SMTPFrom, fromName: cfg.SMTPFromName,
		}
	}
	return developmentVerificationSender{}
}

func (s smtpVerificationSender) SendVerificationCode(_ context.Context, recipient, code string) error {
	port := s.port
	if port == "" {
		port = "587"
	}
	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}
	message := strings.Join([]string{
		"From: " + s.headerFrom(),
		"To: " + recipient,
		"Subject: Ma xac thuc email Finora",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		fmt.Sprintf("Ma xac thuc Finora cua ban la: %s", code),
		"Ma co hieu luc trong 15 phut. Khong chia se ma nay voi bat ky ai.",
	}, "\r\n")
	return smtp.SendMail(s.host+":"+port, auth, s.from, []string{recipient}, []byte(message))
}

func (s smtpVerificationSender) headerFrom() string {
	name := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s.fromName, "\r", ""), "\n", ""))
	if name == "" {
		return s.from
	}
	return fmt.Sprintf("%s <%s>", name, s.from)
}

func (developmentVerificationSender) SendVerificationCode(_ context.Context, recipient, code string) error {
	// Local development has no mail transport. Never enable this mode in a
	// production environment: Config.Load rejects that configuration.
	log.Printf("development email verification code for %s: %s", recipient, code)
	return nil
}
