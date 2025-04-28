package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
)

var db = make(map[string]string)

type GithubClientInterface interface {
	DownloadAsset(url string, outputPath string)
}
type GithubClient struct {
	AccessToken string
	HTTPClient  *http.Client
}

func (c *GithubClient) DownloadAsset(url string, outputPath string) {
	// create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// set the authorization header with the access token
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	// set the accept header to application/octet-stream
	req.Header.Set("Accept", "application/octet-stream")

	fmt.Println("Downloading asset from:", url)

	// perform the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Println("Error performing request:", err.Error())
		return
	}
	defer resp.Body.Close()

	// check if the request was successful
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to perform request:", resp.StatusCode)
		return
	}

	// create a file to save the downloaded asset
	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}
	fmt.Println("Asset downloaded successfully")
}

func checkExistsAndCreateFolder(folderPath string) (string, error) {
	// check if the folder path is empty
	if folderPath == "" {
		return folderPath, errors.New("folder path is empty")
	}

	// check if the folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// if folder does not exist, create it
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			return folderPath, err
		}

		return folderPath, nil
	}

	return folderPath, nil
}

// checks if the assets folder exists, if not creates it
func checkAndCreateAssetsFolder(assetsFolder string) (string, error) {
	return checkExistsAndCreateFolder(assetsFolder)
}

// checks if the owner folder exists, if not creates it
func checkAndCreateOwnerFolder(assetsFolder string, owner string) (string, error) {
	_, assetsFolderCreationError := checkAndCreateAssetsFolder(assetsFolder)
	if assetsFolderCreationError != nil {
		return "", assetsFolderCreationError
	}

	ownerFolder := filepath.Join(assetsFolder, owner)
	_, repositoryFolderCreationError := checkExistsAndCreateFolder(ownerFolder)
	if repositoryFolderCreationError != nil {
		return "", repositoryFolderCreationError
	}

	return ownerFolder, nil
}

// checks if the repository folder exists, if not creates it
func checkAndCreateRepositoryFolder(assetsFolder string, owner string, repo string) (string, error) {
	_, ownerFolderCreationErr := checkAndCreateOwnerFolder(assetsFolder, owner)
	if ownerFolderCreationErr != nil {
		return "", ownerFolderCreationErr
	}

	repositoryFolder := filepath.Join(assetsFolder, owner, repo)
	_, repositoryFolderCreationError := checkExistsAndCreateFolder(repositoryFolder)
	if repositoryFolderCreationError != nil {
		return "", repositoryFolderCreationError
	}

	return repositoryFolder, nil
}

// checks if the release folder exists, if not creates it
func checkAndCreateReleaseFolder(assetsFolder string, owner string, repo string, tag string) (string, error) {
	_, repositoryFolderCreationErr := checkAndCreateRepositoryFolder(assetsFolder, owner, repo)
	if repositoryFolderCreationErr != nil {
		return "", repositoryFolderCreationErr
	}

	releaseFolder := filepath.Join(assetsFolder, owner, repo, tag)
	_, releaseFolderCreationErr := checkExistsAndCreateFolder(releaseFolder)
	if releaseFolderCreationErr != nil {
		return "", releaseFolderCreationErr
	}

	return releaseFolder, nil
}

type RouterOptions struct {
	AssetsFolder string
	GithubClient GithubClientInterface
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
			// TODO(cemreyavuz): check if action is "published"

			releaseFolder, err := checkAndCreateReleaseFolder(routerOptions.AssetsFolder, *event.Repo.Owner.Login, *event.Repo.Name, *event.Release.TagName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create release folder"})
				return
			}

			// TODO(cemreyavuz): check if the asset is already downloaded
			// TODO(cemreyavuz): check if there are more than one asset

			// download the asset
			assetPath := filepath.Join(releaseFolder, *event.Release.Assets[0].Name)
			routerOptions.GithubClient.DownloadAsset(*event.Release.Assets[0].URL, assetPath)

			// TODO(cemreyavuz): unzip the asset

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

func main() {
	fmt.Println("Initializing deploy-to-vm")

	// read assets folder from env
	assetsFolder := os.Getenv("DEPLOY_TO_VM_ASSETS_FOLDER")
	_, err := checkAndCreateAssetsFolder(assetsFolder)
	if err != nil {
		fmt.Println("Error creating assets folder:", err)
		return
	}

	fmt.Println("Assets folder:", assetsFolder)

	// read github access token from env
	githubAccessToken := os.Getenv("GITHUB_TOKEN")
	githubClient := &GithubClient{
		AccessToken: githubAccessToken,
		HTTPClient:  &http.Client{},
	}

	r := setupRouter(RouterOptions{AssetsFolder: assetsFolder, GithubClient: githubClient})
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
