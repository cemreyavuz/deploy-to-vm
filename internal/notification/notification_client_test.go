package notification

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadWebhookUrl_Success(t *testing.T) {
	// Arrange: set the environment variable
	t.Setenv("DEPLOY_TO_VM_NOTIFICATION_WEBHOOK_URL", "http://example.com/webhook")

	// Act: call LoadWebhookUrl
	client := &NotificationClient{}
	err := client.LoadWebhookUrl()

	// Assert: check if the webhook URL is set correctly
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/webhook", client.WebhookURL)
}

func TestLoadWebhookUrl_MissingEnv(t *testing.T) {
	// Arrange: unset the environment variable
	os.Unsetenv("DEPLOY_TO_VM_NOTIFICATION_WEBHOOK_URL")

	// Act: call LoadWebhookUrl
	client := &NotificationClient{}
	err := client.LoadWebhookUrl()

	// Assert: check if the error is returned
	assert.Error(t, err)
}

func TestNotify_Success(t *testing.T) {
	// Arrange: create a test server to mock the webhook URL
	received := ""
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Act: create a NotificationClient with the test server URL
	client := &NotificationClient{WebhookURL: ts.URL}
	err := client.Notify("hello world")

	// Assert: check if the notification was sent successfully
	assert.NoError(t, err)
	assert.Equal(t, `{"content": "hello world"}`, received)
}

func TestNotify_EmptyMessage(t *testing.T) {
	// Act: call Notify with an empty message
	client := &NotificationClient{WebhookURL: "http://example.com/webhook"}
	err := client.Notify("")

	// Assert: check if the error is returned
	assert.Error(t, err)
}

func TestNotify_NoWebhookURL(t *testing.T) {
	// Act: call Notify without setting the webhook URL
	client := &NotificationClient{WebhookURL: ""}
	err := client.Notify("test message")

	// Assert: check if the notification is skipped without error
	assert.NoError(t, err)
}

func TestNotify_PostError(t *testing.T) {
	// Arrange: create a test server to mock the webhook URL
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // Simulate a bad request
		w.Write([]byte("Bad Request"))
		panic("Simulated error") // Simulate an error in the handler
	}))
	defer ts.Close()

	// Act: create a NotificationClient with the test server URL
	client := &NotificationClient{WebhookURL: ts.URL}
	err := client.Notify("fail message")

	// Assert: check if the error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error sending notification:")
}
