package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/AlexMickh/speak-auth/internal/config"
)

type Vars struct {
	ID   string
	Name string
}

func Send(cfg config.MailConfig, to string, id, name string) error {
	const op = "lib.email.Send"

	auth := smtp.PlainAuth("", cfg.FromAddr, cfg.Password, cfg.Host)

	tmpl, err := template.ParseFiles("./internal/lib/email/templates/email.html")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var rendered bytes.Buffer
	vars := Vars{
		Name: name,
		ID:   id,
	}
	if err = tmpl.Execute(&rendered, vars); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"

	err = smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		auth,
		cfg.FromAddr,
		[]string{to},
		fmt.Appendf(nil, "Subject: Email\n%s\n\n%s", headers, rendered.String()),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
