package notification

import (
	http_utils "deploy-to-vm/internal/http-utils"
	"errors"
	"fmt"
	"log"
	"os"
)

type NotificationClient struct {
	WebhookURL string
}

type NotificationClientInterface interface {
	LoadWebhookUrl() error
	Notify(message string) error
}

func (c *NotificationClient) LoadWebhookUrl() error {
	notificationUrl := os.Getenv("DEPLOY_TO_VM_NOTIFICATION_WEBHOOK_URL")
	if notificationUrl == "" {
		return errors.New("Notification webhook URL is not set in environment variables")
	}

	c.WebhookURL = notificationUrl
	log.Println("Notification webhook URL is loaded")
	return nil
}

func (c *NotificationClient) Notify(message string) error {
	if message == "" {
		return errors.New("message cannot be empty")
	}

	if c.WebhookURL == "" {
		log.Println("Notification webhook URL is not set, skipping notification")
		return nil
	}

	// Prepare the data to be sent in the POST request
	data := []byte(fmt.Sprintf(`{"content": "%s"}`, message))

	// Make the POST request to the notification webhook URL
	_, postErr := http_utils.MakePostRequest(c.WebhookURL, data)
	if postErr != nil {
		return errors.New("Error sending notification: " + postErr.Error())
	}

	log.Println("Notification sent:", message)
	return nil
}

func SetupNotificationClient() *NotificationClient {
	notificationClient := &NotificationClient{}
	notificationClient.LoadWebhookUrl()
	if notificationClient.WebhookURL == "" {
		log.Println("Notification webhook URL is not set, notifications will not be sent")
	} else {
		log.Println("Notification webhook URL is set, notifications will be sent")
	}
	return notificationClient
}
