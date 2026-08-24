// Mailer sends OTP codes to the user's email.
//
// Two modes are supported:
//
//   - "console" (default): print the OTP to stdout. This is the most
//     useful mode for local development and the demo CI because it
//     doesn't require an SMTP server.
//   - "smtp": send via net/smtp using the host/port/username/password
//     fields from the [smtp] section of config.toml.
package demoapp

import (
	"errors"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

// Mailer is the interface implemented by both the console and SMTP senders.
type Mailer interface {
	Send(to, subject, body string) error
}

// NewMailer returns a Mailer according to cfg.SMTP.Mode.
func NewMailer(cfg SMTPConfig) (Mailer, error) {
	switch strings.ToLower(cfg.Mode) {
	case "", "console":
		return &ConsoleMailer{}, nil
	case "smtp":
		return &SMTPMailer{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unknown smtp.mode %q", cfg.Mode)
	}
}

// ConsoleMailer — sends by printing to stdout. Intended for dev/demo.
type ConsoleMailer struct{}

// Send prints "To / Subject / Body" on three lines, prefixed so the line
// is easy to grep in container logs.
func (m *ConsoleMailer) Send(to, subject, body string) error {
	if to == "" {
		return errors.New("console mailer: empty recipient")
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	fmt.Fprintf(os.Stdout, "[demoapp-otp] to=%s subject=%q at=%s\n%s\n", to, subject, ts, body)
	return nil
}

// SMTPMailer delivers messages through a real SMTP server.
type SMTPMailer struct {
	cfg SMTPConfig
}

// Send connects to the SMTP server and submits the message. The envelope
// sender is the bare address (no display name) because most servers reject
// the "Name <addr>" form as a MAIL FROM argument with 501.
func (m *SMTPMailer) Send(to, subject, body string) error {
	cfg := m.cfg
	if cfg.Host == "" {
		return errors.New("smtp.host not configured")
	}
	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	fromRaw := cfg.From
	if fromRaw == "" {
		fromRaw = "noreply@opennhp.org"
	}
	fromAddr, err := mail.ParseAddress(fromRaw)
	if err != nil {
		return fmt.Errorf("invalid smtp.from: %w", err)
	}

	// Build RFC-822 message. Use CRLF line endings as required by SMTP.
	if subject == "" {
		subject = cfg.Subject
	}
	if subject == "" {
		subject = "Your OpenNHP Demo verification code"
	}
	codeInSubject := cfg.CodeInSubject == nil || *cfg.CodeInSubject

	// We can't see whether `subject` already includes the OTP — we let the
	// caller pass the fully-rendered subject so this layer doesn't need
	// to know the OTP value. The caller in handler_nhp.go is responsible
	// for appending the code if CodeInSubject is true.
	if codeInSubject {
		// Normalize: collapse any trailing ":" so we don't end up with "::".
		subject = strings.TrimRight(subject, ": ")
	}

	domain := "opennhp.org"
	if at := strings.LastIndex(fromAddr.Address, "@"); at >= 0 && at+1 < len(fromAddr.Address) {
		domain = fromAddr.Address[at+1:]
	}
	now := time.Now()
	msgID := fmt.Sprintf("<%d.%d@%s>", now.UnixNano(), os.Getpid(), domain)

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", fromAddr.String())
	fmt.Fprintf(&b, "To: %s\r\n", toAddr.String())
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", subject))
	fmt.Fprintf(&b, "Date: %s\r\n", now.Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Message-ID: %s\r\n", msgID)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	b.WriteString("\r\n")

	port := cfg.Port
	if port <= 0 {
		port = 587
	}
	addr := cfg.Host + ":" + strconv.Itoa(port)
	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	return smtp.SendMail(addr, auth, fromAddr.Address, []string{toAddr.Address}, []byte(b.String()))
}
