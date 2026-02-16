package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"superaib/internal/storage/repo"

	"gopkg.in/gomail.v2"
)

// ✅ 1. Interface-ka (Kani waa kan xallinaya 'undefined: EmailService')
type EmailService interface {
	SendEmail(ctx context.Context, projectID, toEmail, subject, body string) error
}

type emailService struct {
	configRepo repo.ProjectAuthConfigRepository
}

// Constructor
func NewEmailService(cr repo.ProjectAuthConfigRepository) EmailService {
	return &emailService{configRepo: cr}
}

// Implementation
func (s *emailService) SendEmail(ctx context.Context, projectID, toEmail, subject, body string) error {
	// 1. Soo saar SMTP Config-ga Developer-ka (oo magaciisu yahay 'smtp_email')
	// Magacan 'smtp_email' waa kii aad Admin Panel-ka ku soo qortay
	config, err := s.configRepo.GetByProjectAndProviderName(ctx, projectID, "smtp_email")
	if err != nil {
		return errors.New("SMTP settings not configured for this project")
	}

	// 2. Kala furfur JSON-ka (Credentials)
	var creds map[string]interface{}
	if err := json.Unmarshal(config.Credentials, &creds); err != nil {
		return fmt.Errorf("invalid smtp config format: %w", err)
	}

	// 3. Hel xogta lagama maarmaanka ah (Safetly cast to string)
	host, _ := creds["host"].(string)
	username, _ := creds["username"].(string)
	password, _ := creds["password"].(string)
	senderEmail, _ := creds["sender_email"].(string)
	senderName, _ := creds["sender_name"].(string)

	// Port-ka wuxuu mararka qaar noqon karaa string ama number, labadaba waan eegaynaa
	var port int
	if p, ok := creds["port"].(string); ok {
		port, _ = strconv.Atoi(p)
	} else if p, ok := creds["port"].(float64); ok {
		port = int(p)
	}

	// Validation
	if host == "" || port == 0 || username == "" || password == "" {
		return errors.New("incomplete SMTP configuration: missing host, port, username, or password")
	}

	// 4. Diyaari Email-ka (Gomail Library)
	m := gomail.NewMessage()

	// Set Sender (e.g. "SuperAIB <support@mysite.com>")
	if senderName != "" {
		m.SetHeader("From", fmt.Sprintf("%s <%s>", senderName, senderEmail))
	} else {
		m.SetHeader("From", senderEmail)
	}

	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body) // Waxaan u diraynaa HTML ahaan

	// 5. Connect and Send
	d := gomail.NewDialer(host, port, username, password)

	// Haddii aad Zoho isticmaalayso, inta badan TLS way u baahan yihiin
	// Gomail automatically ayuu u maareeyaa, laakiin hubi in port 465 (SSL) ama 587 (TLS) uu sax yahay.

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}
