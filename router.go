package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
)

type RouterOptions struct {
	AssetsDir    string
	ConfigClient ConfigClientInterface
	GithubClient GithubClientInterface
	NginxClient  NginxClientInterface
}

func setupRouter(routerOptions RouterOptions) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.POST("/deploy-with-gh", func(c *gin.Context) {
		// TODO(cemreyavuz): setup a secret token for the webhook
		// validate payload
		payload, err := github.ValidatePayload(c.Request, nil)
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
			releaseDir, createReleaseDirErr := createReleaseDirIfIsNotExist(
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

			// Link release assets to site directory
			// TODO(cemreyavuz): read site dir from DB
			siteDir := os.Getenv("DEPLOY_TO_VM_SITE_DIR")
			moveErr := linkReleaseAssetsToSiteDir(releaseDir, siteDir)
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

			c.JSON(http.StatusOK, gin.H{"action": *event.Action})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported event type"})
		}
	})

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}
