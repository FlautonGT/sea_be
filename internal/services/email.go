package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

// EmailService handles sending emails
type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromName     string
	FromEmail    string
	AppURL       string
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost:     getEnv("SMTP_HOST", "smtp-relay.brevo.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromName:     getEnv("SMTP_FROM_NAME", "Seaply"),
		FromEmail:    getEnv("SMTP_FROM_EMAIL", "noreply@gate.co.id"),
		AppURL:       getEnv("APP_URL", "http://localhost:3000"),
	}
}

// SendVerificationEmail sends email verification link
func (e *EmailService) SendVerificationEmail(to, firstName, verificationToken string) error {
	verificationURL := fmt.Sprintf("%s/verify-email/%s", e.AppURL, verificationToken)

	subject := "Verifikasi Email Anda - Seaply"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
            <h1 style="color: white; margin: 0; font-size: 28px;">Seaply</h1>
            <p style="color: white; margin: 10px 0 0 0; opacity: 0.9;">Top Up Game & Voucher Digital Terpercaya</p>
        </div>

        <div style="background: white; padding: 40px 30px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px;">
            <h2 style="color: #1f2937; margin-top: 0;">Halo, %s!</h2>

            <p style="color: #4b5563; font-size: 16px; line-height: 1.6;">
                Terima kasih telah mendaftar di Seaply. Untuk melanjutkan, silakan verifikasi email Anda dengan mengklik tombol di bawah ini:
            </p>

            <div style="text-align: center; margin: 35px 0;">
                <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 14px 40px; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 16px;">
                    Verifikasi Email
                </a>
            </div>

            <p style="color: #6b7280; font-size: 14px; line-height: 1.6;">
                Atau salin dan tempel link berikut di browser Anda:
            </p>

            <div style="background: #f9fafb; padding: 15px; border-radius: 6px; margin: 15px 0; word-break: break-all;">
                <a href="%s" style="color: #667eea; text-decoration: none; font-size: 14px;">%s</a>
            </div>

            <p style="color: #6b7280; font-size: 14px; line-height: 1.6; margin-top: 30px;">
                Link verifikasi ini akan kedaluwarsa dalam 30 menit.
            </p>

            <p style="color: #6b7280; font-size: 14px; line-height: 1.6;">
                Jika Anda tidak membuat akun di Seaply, abaikan email ini.
            </p>
        </div>

        <div style="text-align: center; padding: 20px; color: #9ca3af; font-size: 12px;">
            <p style="margin: 5px 0;">&copy; 2025 Seaply. All rights reserved.</p>
            <p style="margin: 5px 0;">
                <a href="https://gate.co.id" style="color: #667eea; text-decoration: none;">Website</a> |
                <a href="https://gate.co.id/terms" style="color: #667eea; text-decoration: none;">Terms</a> |
                <a href="https://gate.co.id/privacy" style="color: #667eea; text-decoration: none;">Privacy</a>
            </p>
        </div>
    </div>
</body>
</html>
	`, firstName, verificationURL, verificationURL, verificationURL)

	return e.send(to, subject, htmlBody)
}

// SendPasswordResetEmail sends password reset link
func (e *EmailService) SendPasswordResetEmail(to, firstName, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password/%s", e.AppURL, resetToken)

	subject := "Reset Password - Seaply"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
            <h1 style="color: white; margin: 0; font-size: 28px;">Seaply</h1>
            <p style="color: white; margin: 10px 0 0 0; opacity: 0.9;">Top Up Game & Voucher Digital Terpercaya</p>
        </div>

        <div style="background: white; padding: 40px 30px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px;">
            <h2 style="color: #1f2937; margin-top: 0;">Halo, %s!</h2>

            <p style="color: #4b5563; font-size: 16px; line-height: 1.6;">
                Anda telah meminta untuk mereset password akun Seaply Anda. Klik tombol di bawah ini untuk melanjutkan:
            </p>

            <div style="text-align: center; margin: 35px 0;">
                <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 14px 40px; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 16px;">
                    Reset Password
                </a>
            </div>

            <p style="color: #6b7280; font-size: 14px; line-height: 1.6;">
                Atau salin dan tempel link berikut di browser Anda:
            </p>

            <div style="background: #f9fafb; padding: 15px; border-radius: 6px; margin: 15px 0; word-break: break-all;">
                <a href="%s" style="color: #667eea; text-decoration: none; font-size: 14px;">%s</a>
            </div>

            <p style="color: #ef4444; font-size: 14px; line-height: 1.6; margin-top: 30px; padding: 12px; background: #fef2f2; border-radius: 6px;">
                ⚠️ Link ini akan kedaluwarsa dalam 1 jam. Jika Anda tidak meminta reset password, abaikan email ini.
            </p>
        </div>

        <div style="text-align: center; padding: 20px; color: #9ca3af; font-size: 12px;">
            <p style="margin: 5px 0;">&copy; 2025 Seaply. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
	`, firstName, resetURL, resetURL, resetURL)

	return e.send(to, subject, htmlBody)
}

// send sends an email using SMTP
func (e *EmailService) send(to, subject, htmlBody string) error {
	// Create message
	from := fmt.Sprintf("%s <%s>", e.FromName, e.FromEmail)

	// Build email headers and body
	message := bytes.NewBuffer(nil)
	message.WriteString(fmt.Sprintf("From: %s\r\n", from))
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)

	// SMTP authentication
	auth := smtp.PlainAuth("", e.SMTPUser, e.SMTPPassword, e.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)
	err := smtp.SendMail(addr, auth, e.FromEmail, []string{to}, message.Bytes())

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// RenderTemplate renders an email template (for future use)
func RenderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, err := template.New(templateName).Parse(emailTemplates[templateName])
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Email templates (can be moved to separate files later)
var emailTemplates = map[string]string{
	"verification": `...`,
	"reset":        `...`,
}
