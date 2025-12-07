package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"os"
)

// SendEmail sends an HTML email message to the specified recipient.
// This function provides a complete email delivery workflow:
// 1. Configures email headers including subject and MIME type
// 2. Sets up a secure TLS connection to the SMTP server
// 3. Authenticates with the server using credentials from environment variables
// 4. Sends the HTML-formatted message
// 5. Properly closes all connections
//
// The function uses environment variables for configuration:
// - EMAIL: Sender's email address
// - EMAIL_PASSWORD: Sender's email password
// - EMAIL_HOST: SMTP server hostname
// - EMAIL_PORT: SMTP server port
//
// Parameters:
//   - email: The recipient's email address
//   - subject: The email subject line
//   - body: The HTML content of the email body
//
// Returns: Error if any step in the email sending process fails, with context about which step failed
func SendEmail(email string, subject string, body string) error {
	subject = "Subject: " + subject + "\n"
	from := os.Getenv("EMAIL")
	pass := os.Getenv("EMAIL_PASSWORD")
	host := os.Getenv("EMAIL_HOST")
	port := os.Getenv("EMAIL_PORT")
	serverName := host + ":" + port

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	message := []byte(subject + mime + body)

	tlsconfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false, // Enforces certificate validation for security
	}

	// Establish TCP connection to the SMTP server
	conn, err := net.Dial("tcp", serverName)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}

	defer conn.Close()

	// Create a new SMTP client
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("client creation error: %w", err)
	}
	defer client.Close()

	// Identify ourselves to the SMTP server
	if err = client.Hello("localhost"); err != nil {
		return fmt.Errorf("HELO error: %w", err)
	}

	// Check if the server supports STARTTLS
	ok, _ := client.Extension("STARTTLS")
	if !ok {
		return fmt.Errorf("extension error: %w", err)
	}

	// Upgrade to TLS connection
	if err = client.StartTLS(tlsconfig); err != nil {
		return fmt.Errorf("TLS error: %w", err)
	}

	// Authenticate with the server
	auth := smtp.PlainAuth("", from, pass, host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication error: %w", err)
	}

	// Set the sender
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM error: %w", err)
	}

	// Set the recipient
	if err = client.Rcpt(email); err != nil {
		return fmt.Errorf("RCPT TO error: %w", err)
	}

	// Get a writer from the client
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command error: %w", err)
	}

	// Write the message
	if _, err = w.Write(message); err != nil {
		return fmt.Errorf("message write error: %w", err)
	}

	// Close the writer
	if err = w.Close(); err != nil {
		return fmt.Errorf("writer close error: %w", err)
	}

	// Quit the session
	if err = client.Quit(); err != nil {
		return fmt.Errorf("QUIT command error: %w", err)
	}

	return nil
}

// RenderTemplate renders an HTML template with the provided data.
// This function performs template processing in two steps:
// 1. Parses the HTML template file from the specified path
// 2. Executes the template with the provided data, injecting values into placeholders
//
// The function is designed to work with Go's standard template package, supporting
// all template features including conditionals, loops, and custom functions.
//
// Parameters:
//   - tpl: The filesystem path to the HTML template file
//   - data: The data to inject into the template (can be any Go value: map, struct, etc.)
//
// Returns:
//   - The rendered HTML content as a string
//   - Error if template parsing or execution fails
func RenderTemplate(tpl string, data any) (string, error) {
	// Parse the template from the file
	tmpl, err := template.ParseFiles(tpl)
	if err != nil {
		return "", fmt.Errorf("template parsing error: %w", err)
	}

	// Create a buffer to store the rendered output
	var buf bytes.Buffer

	// Execute the template, injecting the provided data
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	// Return the rendered content as a string
	return buf.String(), nil
}

