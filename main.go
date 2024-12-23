package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mailgun/mailgun-go/v4"
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func SendPasswordRecoveryMailWithOTP(domain, apiKey, recipient, reset_otp string) (string, error) {
	// Initialize Mailgun client
	mg := mailgun.NewMailgun(domain, apiKey)

	// Generate password recovery URL
	// resetURL := fmt.Sprintf(constants.PASSWORD_RECOVERY", token)

	// Create the email message
	m := mg.NewMessage(
		"Support <no-reply@mail.conneft.com>",                             // Sender email
		"Password Recovery",                                               // Subject
		fmt.Sprintf("Copy the OTP to reset your password: %s", reset_otp), // Plain text body (optional)
		recipient, // Recipient email
	)
	username, err := extractUsername(recipient)
	if err != nil {
		return "", err
	}
	// Include HTML content
	m.SetHTML(fmt.Sprintf(`
		<h1>Password Recovery</h1>
		<p>Hello %s,</p>
		<p>We have received a request to reset your password.</p>
		<p>Copy the OTP below to reset your password:</p>
		<p>%s</p>
		<p>Link expires in 5 minutes.</p>
		<p>If you did not request a password reset, please ignore this email.</p>
	`, username, reset_otp))

	// Set a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the email
	_, id, err := mg.Send(ctx, m)
	return id, err
}

func SendWaitlistWelcomeEmail(domain, apiKey, recipient, username string) (string, error) {
	mg := mailgun.NewMailgun(domain, apiKey)

	m := mailgun.NewMessage(
		"Conneft Team <no-reply@mail.conneft.com>",
		"You're In! Here's How to Stay Connected with Conneft 🚀",
		"",
		recipient,
	)

	// Get username from email
	// firstName, err := extractUsername(recipient)
	// if err != nil {
	// 	return "", err
	// }

	// Open and attach the images
	headerFile, err := os.Open("./header.jpg")
	if err != nil {
		return "", fmt.Errorf("failed to open header image: %v", err)
	}
	defer headerFile.Close()

	footerFile, err := os.Open("./footer.jpg")
	if err != nil {
		return "", fmt.Errorf("failed to open footer image: %v", err)
	}
	defer footerFile.Close()

	// Add inline images with Content-IDs
	m.AddReaderInline("header", headerFile)
	m.AddReaderInline("footer", footerFile)

	// Reference the images in HTML using cid:
	m.SetHTML(fmt.Sprintf(`
        <div>
            <img src="cid:header" alt="Header Image" style="width: 100%%;" />
            <h1>Welcome to Conneft! 🎉</h1>
            <p>Hello %s,</p>
            <p>You've taken the first step toward a new era of networking — where physical and digital connections meet effortlessly on one platform. We're thrilled to have you on this journey as we redefine how meaningful connections are made.</p>
            <h2>Here's What's Coming Your Way:</h2>
            <ul>
                <li>✨ <b>Tap-to-Connect Simplicity:</b> Share who you are and what you do effortlessly.</li>
                <li>🌟 <b>Build Your Social Presence:</b> Stand out in any room — digitally and physically.</li>
                <li>🚀 <b>Early Access:</b> Be the first to experience cutting-edge features built for the future of networking.</li>
            </ul>
            <p>Stay tuned for exciting updates, early access offers, and sneak peeks leading up to our launch. You're part of something big, and we can't wait to share it all with you.</p>
            <p><b>Quick Tip:</b> If you find this email in your spam or promotions folder, kindly mark us as "Not Spam" and move us to your inbox. This ensures you won't miss out on exclusive updates!</p>
            <p>Thank you for being part of the Conneft community. The future of connection starts here!</p>
            <p>Warm Regards,</p>
            <p><b>The Conneft Team</b></p>
            <img src="cid:footer" alt="Footer Image" style="width: 100%%;" />
        </div>
    `, username))

	// Set a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Send the email
	_, id, err := mg.Send(ctx, m)
	return id, err
}
func extractUsername(email string) (string, error) {
	// Split the email into username and domain parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid email address: %s", email)
	}
	return parts[0], nil
}

// handlers

func SendWaitlistWelcomeEmailHandler(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		log.Println("email is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	username := c.Param("username")
	if username == ""{
		log.Println("username is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	apiKey := os.Getenv("MAILGUN_API_KEY")
	if apiKey == "" {
		log.Println("MAILGUN_API_KEY is not set or empty")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	// Send the welcome email
	id, err := SendWaitlistWelcomeEmail("mail.conneft.com", apiKey, email, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email sent", "id": id})
}

func main() {
	// apiKey := os.Getenv("MAILGUN_API_KEY")
	// if apiKey == "" {
	// 	log.Fatalf("MAILGUN_API_KEY is not set or empty. Check your .env file or environment variables.")
	// }

	// log.Printf("Loaded MAILGUN_API_KEY: %s", apiKey)

	// id, err := SendWaitlistWelcomeEmail("mail.conneft.com", apiKey, "jasonabbatheleast@gmail.com")
	// if err != nil {
	// 	log.Fatalf("Error sending email: %v", err)
	// }

	// log.Printf("Email sent with ID: %s", id)
	

	// Initialize the Gin router
	r := gin.Default()
	r.POST("/waitlist/send-email/:email/:username", SendWaitlistWelcomeEmailHandler)
	r.Run(":8073")
}
