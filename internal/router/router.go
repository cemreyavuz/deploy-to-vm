package router

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"deploy-to-vm/internal/config"
	file_utils "deploy-to-vm/internal/file-utils"
	deploy_to_vm_github "deploy-to-vm/internal/github"
	"deploy-to-vm/internal/nginx"
	"deploy-to-vm/internal/notification"
	"deploy-to-vm/internal/pm2"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
)

type RouterOptions struct {
	AssetsDir          string
	ConfigClient       config.ConfigClientInterface
	GithubClient       deploy_to_vm_github.GithubClientInterface
	NginxClient        nginx.NginxClientInterface
	NotificationClient notification.NotificationClientInterface
	Pm2Client          pm2.Pm2ClientInterface
	SecretToken        string
}

func SetupRouter(routerOptions RouterOptions) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.POST("/deploy-with-gh", func(c *gin.Context) {
		// validate payload
		var (
			payload       []byte
			validationErr error
		)
		if routerOptions.ConfigClient.IsDevelopment() {
			log.Println("Developmet mode is enabled, not validating signature")
			payload, validationErr = github.ValidatePayload(c.Request, nil)
		} else {
			payload, validationErr = github.ValidatePayload(c.Request, []byte(routerOptions.SecretToken))
		}
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid payload %v", validationErr)})
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
			code, downloadErr := routerOptions.GithubClient.DownloadAssets(event.Release.Assets, releaseDir)
			if downloadErr != nil {
				// TODO(cemreyavuz): return a different error code depending on the error
				switch code {
				case deploy_to_vm_github.DownloadAsset_NoAssetsFound:
					log.Printf("No assets found for release: \"%s\", will skip the request.", *event.Release.TagName)
					c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("No assets found for release \"%s\", will skip the request.", *event.Release.TagName)})
					return
				default:
					log.Printf("Failed to download assets: \"%v\"", downloadErr)
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to download assets: %v", downloadErr)})
				}
				// TODO(cemreyavuz): remove the release directory if download fails
				return
			}

			// Untar files in the release directory
			files, untarErr := file_utils.UntarGzFilesInDir(releaseDir)
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

			// Reload the target service (nginx or pm2)
			reloadFn := func() error {
				return fmt.Errorf("reload function is not defined for targetType: %s", repositoryConfig.TargetType)
			}

			if repositoryConfig.TargetType == "nginx" {
				reloadFn = routerOptions.NginxClient.Reload
			} else if repositoryConfig.TargetType == "pm2" {
				reloadFn = func() error {
					return routerOptions.Pm2Client.Reload(repositoryConfig.TargetProcessName)
				}
			}

			reloadErr := reloadFn()
			if reloadErr != nil {
				log.Printf("Failed to reload nginx unit: \"%v\"", reloadErr)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload nginx unit"})
				return
			}

			// Send notification
			notificationMessage := fmt.Sprintf("New release deployed for: `repo:%s` `tag:%s`\\n\\nFiles:\\n```\\n- %s\\n```", *event.Repo.Name, *event.Release.TagName, strings.Join(files, "\\n- "))
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
