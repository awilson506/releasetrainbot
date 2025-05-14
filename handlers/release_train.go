package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/awilson506/releasetrain/db"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

// SlackCommandRequest slash command request structure
type SlackCommandRequest struct {
	Token               string `form:"token"`
	TeamID              string `form:"team_id"`
	TeamDomain          string `form:"team_domain"`
	ChannelID           string `form:"channel_id" example:"CTLGNDNV9"`
	ChannelName         string `form:"channel_name"`
	UserID              string `form:"user_id" example:"U123456"`
	UserName            string `form:"user_name"`
	Command             string `form:"command" example:"/release-train"`
	Text                string `form:"text"`
	APIAppID            string `form:"api_app_id"`
	IsEnterpriseInstall string `form:"is_enterprise_install"`
	ResponseURL         string `form:"response_url"`
	TriggerID           string `form:"trigger_id"`
}

// @Summary Update or create a channel rotation
// @Description Updates the rotation for a channel or creates a new rotation if it doesn't exist
// @Tags slack
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param        token           formData string false  "Slack verification token"
// @Param        team_id         formData string false  "Slack Team ID"
// @Param        team_domain     formData string false  "Slack Team Domain"
// @Param        enterprise_id   formData string false "Slack Enterprise ID"
// @Param        enterprise_name formData string false "Slack Enterprise Name"
// @Param        channel_id      formData string true  "Channel ID where command was issued"
// @Param        channel_name    formData string false  "Channel name"
// @Param        user_id         formData string false  "User ID of the person who triggered the command"
// @Param        user_name       formData string false  "Username of the person who triggered the command"
// @Param        command         formData string true  "Slash command (e.g., /release-train)"
// @Param        text            formData string false "Text following the command"
// @Param        response_url    formData string false  "URL to send delayed response to"
// @Param        trigger_id      formData string false  "Trigger ID for interactive components"
// @Param        api_app_id      formData string false  "Slack App ID"
// @Success 200 {string} string "Rotation updated"
// @Failure 400 {string} string "Bad Request"
// @Router /slack/command [post]
func SlackReleaseTrainHandler(c *gin.Context) {
	var request SlackCommandRequest

	_ = c.Bind(&request)
	text := strings.TrimSpace(request.Text)

	if request.ChannelID == "" {
		response := createSlackMessage(
			slack.ResponseTypeEphemeral,
			slack.MarkdownType,
			":warning: *Missing channel_id in request.*",
		)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	switch {
	case text == "":
		// Fetch and return the current rotation
		rows, err := db.DB.Query(`
			SELECT user_id, position FROM users WHERE channel_id = ? ORDER BY position ASC
		`, request.ChannelID)
		if err != nil {
			response := createSlackMessage(
				slack.ResponseTypeEphemeral,
				slack.MarkdownType,
				":warning: *Failed to fetch rotation.*",
			)
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		defer rows.Close()

		var rotation []string
		for rows.Next() {
			var userID string
			var position int
			if err := rows.Scan(&userID, &position); err == nil {
				rotation = append(rotation, fmt.Sprintf("<@%s>", userID))
			}
		}

		if len(rotation) == 0 {
			response := createSlackMessage(
				slack.ResponseTypeEphemeral,
				slack.MarkdownType,
				":warning: *No rotation set for this channel.*",
			)

			c.JSON(http.StatusOK, response)
			return
		}

		// Join names into a list
		list := ""
		for i, name := range rotation {
			list += fmt.Sprintf("%d. %s\n", i+1, name)
		}

		response := createSlackMessage(
			slack.ResponseTypeInChannel,
			slack.MarkdownType,
			"*Current Rotation:*\n"+list,
		)

		c.JSON(http.StatusOK, response)

	case strings.ToLower(text) == "delete" || strings.ToLower(text) == "--delete":
		// Delete rotation for the channel
		_, err := db.DB.Exec(`DELETE FROM users WHERE channel_id = ?`, request.ChannelID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rotation"})
			return
		}

		response := createSlackMessage(
			slack.ResponseTypeEphemeral,
			slack.MarkdownType,
			":white_check_mark: *Rotation deleted for this channel.*",
		)

		c.JSON(http.StatusOK, response)

	default:
		// Parse and update rotation
		userIDs := parseMentions(text)
		if len(userIDs) == 0 {
			response := createSlackMessage(
				slack.ResponseTypeEphemeral,
				slack.MarkdownType,
				":warning: *No valid user mentions found.*",
			)

			c.JSON(http.StatusOK, response)
			return
		}

		tx, err := db.DB.Begin()
		if err != nil {
			response := createSlackMessage(
				slack.ResponseTypeEphemeral,
				slack.MarkdownType,
				":x: *Internal error: Failed to start DB transaction.*",
			)

			// you read the right, 200 for a 500, but we don't want slack to show an error because we messed up
			c.JSON(http.StatusOK, response)
			return
		}
		defer tx.Rollback()

		for i, uid := range userIDs {
			_, err := tx.Exec(`
				INSERT INTO users (user_id, position, channel_id)
				VALUES (?, ?, ?)
				ON DUPLICATE KEY UPDATE position = VALUES(position)
			`, uid, i, request.ChannelID)

			if err != nil {
				response := createSlackMessage(
					slack.ResponseTypeEphemeral,
					slack.MarkdownType,
					fmt.Sprintf(":x: *Failed to update user `%s`.*", uid),
				)

				c.JSON(http.StatusOK, response)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			response := createSlackMessage(
				slack.ResponseTypeEphemeral,
				slack.MarkdownType,
				":x: *Failed to update rotation.*",
			)
			c.JSON(http.StatusOK, response)
			return
		}

		response := createSlackMessage(
			slack.ResponseTypeInChannel,
			slack.MarkdownType,
			fmt.Sprintf(":white_check_mark: *Rotation updated* for <#%s> with *%d* users.", request.ChannelID, len(userIDs)),
		)

		c.JSON(http.StatusOK, response)
	}
}

// parseMentions utility function to parse user mentions from the text
// It extracts user IDs from the text and returns them as a slice of strings
// Example: "<@U12345>" will be parsed to "U12345"
func parseMentions(text string) []string {
	var userIDs []string
	mentions := strings.Split(text, " ")

	for _, mention := range mentions {
		if strings.HasPrefix(mention, "<@") && strings.HasSuffix(mention, ">") {
			userIDs = append(userIDs, mention[2:len(mention)-1]) // Extract user ID
		}
	}
	return userIDs
}

// createSlackMessage utility function to create and return formatted Slack messages
func createSlackMessage(responseType string, msgType string, text string, additionalBlocks ...slack.Block) gin.H {
	msg := slack.NewBlockMessage(
		slack.NewSectionBlock(
			slack.NewTextBlockObject(
				msgType, // Slack markdown type for styling text
				text,    // The message text to display
				false, false,
			),
			nil, nil,
		),
	)

	// If additional blocks are passed (e.g., for attachments or extra information), add them
	if len(additionalBlocks) > 0 {
		msg.Blocks.BlockSet = append(msg.Blocks.BlockSet, additionalBlocks...)
	}

	return gin.H{
		"response_type": responseType,
		"blocks":        msg.Blocks.BlockSet,
	}
}
