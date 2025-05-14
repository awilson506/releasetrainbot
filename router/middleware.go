package routes

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/awilson506/releasetrain/config"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

// simpleMiddleware logs the start and end of a request.
func slackMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("handaling command request")
		if config.IsProduction() {
			err := verifySlackSignature(c.Request)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		// Process request
		c.Next()

		fmt.Println("command request completed")
	}
}

// verifySlackSignature verifies the request signature from Slack
func verifySlackSignature(request *http.Request) (err error) {
	sv, err := slack.NewSecretsVerifier(request.Header, config.SlackSigningKey)
	if err != nil {
		log.Printf("Error creating secrets verifier: %v", err)
		return
	}

	// Read and duplicate body to verify and restore it
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		return
	}
	request.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // reset for downstream

	// Write body into verifier
	if _, err = sv.Write(bodyBytes); err != nil {
		log.Printf("Error writing body to verifier: %v", err)
		return
	}

	// Verify signature
	if err = sv.Ensure(); err != nil {
		log.Printf("Signature verification failed: %v", err)
		return
	}
	return err
}

// checkCloudFrontToken is a middleware that checks for a custom header
// "X-Header-Token" in the request. If the token is not present or does not match
// the expected value, it aborts the request with a 401 Unauthorized status
func checkCloudFrontToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the custom header from the request
		token := c.GetHeader("X-Header-Token")
		if token == "" || token != config.CloudfrontToken {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Next()
	}
}
