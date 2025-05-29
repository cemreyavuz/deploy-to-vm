package main

import (
	"fmt"
	"log"
	"net/http"

	"deploy-to-vm/internal/config"
	file_utils "deploy-to-vm/internal/file-utils"
	deploy_to_vm_github "deploy-to-vm/internal/github"
	"deploy-to-vm/internal/nginx"
	"deploy-to-vm/internal/notification"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
)

type RouterOptions struct {
	AssetsDir          string
	ConfigClient       config.ConfigClientInterface
	GithubClient       deploy_to_vm_github.GithubClientInterface
	NginxClient        nginx.NginxClientInterface
	NotificationClient notification.NotificationClientInterface
	SecretToken        string
}

func setupRouter(routerOptions RouterOptions) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.POST("/deploy-with-gh", func(c *gin.Context) {
		// validate payload
		payload, err := github.ValidatePayload(c.Request, []byte(routerOptions.SecretToken))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		// parse the payload
		event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unrecognized event type"})
			return
		}

		switch event := event.(type) {
		case *github.ReleaseEvent:
			// TODO(cemreyavuz): check if release event has required fields
			if *event.Action != "released" {
				c.JSON(http.StatusOK, gin.H{"message": "Only \"released\" action is supported, ignoring..."})
				return
			}

			// Create release directory if it doesn't exist
			releaseDir, createReleaseDirErr := file_utils.CreateReleaseDirIfIsNotExist(
				routerOptions.AssetsDir,
				*event.Repo.Owner.Login,
				*event.Repo.Name,
				*event.Release.TagName,
			)
			if createReleaseDirErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create release directory"})
				return
			}

			// Download assets
			downloadErr := routerOptions.GithubClient.DownloadAssets(event.Release.Assets, releaseDir)
			if downloadErr != nil {
				log.Printf("Failed to download assets: \"%v\"", downloadErr)
				// TODO(cemreyavuz): remove the release directory if download fails
				// TODO(cemreyavuz): return a different error code depending on the error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download assets"})
				return
			}

			// Untar files in the release directory
			untarErr := file_utils.UntarGzFilesInDir(releaseDir)
			if untarErr != nil {
				log.Printf("Failed to untar files in release directory: \"%v\"", untarErr)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to untar files in release directory"})
				return
			}

			// Link release assets to site directory
			repositoryConfig := routerOptions.ConfigClient.GetRepository(*event.Repo.Name, *event.Repo.Owner.Login)
			if repositoryConfig == nil {
				log.Printf("Repository not found in config: %s/%s", *event.Repo.Owner.Login, *event.Repo.Name)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Repository not found in config"})
				return
			}

			siteDir := repositoryConfig.TargetDir
			if siteDir == "" {
				log.Printf("Site directory not found for repository: %s/%s", *event.Repo.Owner.Login, *event.Repo.Name)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Site directory not found for repository"})
				return
			}

			moveErr := file_utils.LinkReleaseAssetsToSiteDir(releaseDir, siteDir)
			if moveErr != nil {
				log.Printf("Failed to move release assets to site directory: \"%v\"", moveErr)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move release assets to site directory"})
				return
			}

			// Reload nginx unit
			reloadErr := routerOptions.NginxClient.Reload()
			if reloadErr != nil {
				log.Printf("Failed to reload nginx unit: \"%v\"", reloadErr)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload nginx unit"})
				return
			}

			// Send notification
			notificationMessage := fmt.Sprintf("New release (%s) deployed for %s", *event.Release.TagName, *event.Repo.Name)
			notificationErr := routerOptions.NotificationClient.Notify(notificationMessage)
			if notificationErr != nil {
				log.Printf("Failed to send notification: \"%v\"", notificationErr)
			}

			c.JSON(http.StatusOK, gin.H{"action": *event.Action})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported event type"})
		}
	})

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}
