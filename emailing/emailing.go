package emailing

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
)

// SenderType exists to provide context for plain strings in configuration
type SenderType int

// all known types of ErrorReporter
const (
	Log SenderType = iota
	SendGrid
)

// NewErrorReporter will instantiate an ErrorReporter for a known type
func NewSender(t SenderType, credentials string, logger logrus.FieldLogger) Sender {
	logger = logger.WithField("scope", "NewEmailClient")
	switch t {
	case Log:
		return NewLogSender(logger)
	case SendGrid:
		return NewSendgridSender(credentials)
	default:
		return func(email *mail.SGMailV3) error {
			_, err := fmt.Fprintf(os.Stderr, "send email %#v", email)
			return err
		}
	}
}

func NewLogSender(logger logrus.FieldLogger) Sender {
	return func(email *mail.SGMailV3) error {
		fields := logrus.Fields{
			"template_id": email.TemplateID,
			"from":        email.From,
			"subject":     email.Subject,
			"to":          strings.Join(ToAddresses(email), ", "),
			"content":     Content(email),
		}
		for k, v := range TemplateData(email) {
			fields[k] = v
		}
		logger.WithFields(fields).Info("Sending email...")
		return nil
	}
}

func NewSendgridSender(credentials string) Sender {
	emailClient := sendgrid.NewSendClient(credentials)
	return func(email *mail.SGMailV3) error {
		resp, err := emailClient.Send(email)
		if err != nil {
			return err
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("SendGrid responsed with failure status: %v", resp)
		}
		return nil
	}
}

type Sender func(email *mail.SGMailV3) error

func Send(sender Sender, templateID, to string, from *mail.Email, fields ...interface{}) error {
	m := mail.NewV3Mail()
	m.SetTemplateID(templateID)
	m.SetFrom(from)

	p := mail.NewPersonalization()
	p.AddTos(mail.NewEmail("", to))
	td, err := templateData(fields)
	if err != nil {
		return fmt.Errorf("could form email template data from fields passed: %v", err)
	}
	p.DynamicTemplateData = td

	m.AddPersonalizations(p)

	return sender(m)
}

func ToAddresses(m *mail.SGMailV3) []string {
	var tos []string
	for _, p := range m.Personalizations {
		for _, email := range p.To {
			tos = append(tos, email.Address)
		}
	}
	return tos
}

func TemplateData(m *mail.SGMailV3) map[string]interface{} {
	td := make(map[string]interface{})
	for _, p := range m.Personalizations {
		for key, value := range p.DynamicTemplateData {
			td[key] = value
		}
	}
	return td
}

func Content(m *mail.SGMailV3) string {
	buf := new(bytes.Buffer)
	for _, c := range m.Content {
		buf.WriteString(c.Type)
		buf.WriteString("\n")
		buf.WriteString(c.Value)
		buf.WriteString("\n")
	}
	return buf.String()
}

func templateData(fields []interface{}) (map[string]interface{}, error) {
	if len(fields)%2 != 0 {
		return nil, fmt.Errorf("fields passed to email template must be list of key-value pairs")
	}
	m := make(map[string]interface{}, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			return nil, fmt.Errorf("elements passed as keys must be strings but got %#v in position %d",
				fields[i], i)
		}
		m[key] = fields[i+1]
	}
	return m, nil
}
