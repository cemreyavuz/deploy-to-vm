package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

type NotificationClient struct {
	webhookURL string
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

	c.webhookURL = notificationUrl
	log.Println("Notification webhook URL is loaded")
	return nil
}

func (c *NotificationClient) Notify(message string) error {
	if message == "" {
		return errors.New("message cannot be empty")
	}

	if c.webhookURL == "" {
		loadErr := c.LoadWebhookUrl()
		if loadErr != nil {
			return errors.New("Could not load notification webhook URL, not sending notification")
		}
	}

	// Prepare the data to be sent in the POST request
	data := []byte(fmt.Sprintf(`{"content": "%s"}`, message))

	// Make the POST request to the notification webhook URL
	_, postErr := makePostRequest(c.webhookURL, data)
	if postErr != nil {
		return errors.New("Error sending notification: " + postErr.Error())
	}

	log.Println("Notification sent:", message)
	return nil
}
