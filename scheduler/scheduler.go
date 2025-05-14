package scheduler

import (
	"log"

	"github.com/awilson506/releasetrain/config"
	"github.com/awilson506/releasetrain/db"

	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
)

// User user in the rotation system
type User struct {
	UserID   string
	Position int
}

// Start initializes the scheduler and starts the cron job
// to announce weekly rotations.
func Start() {
	c := cron.New()
	c.AddFunc("CRON_TZ=UTC 0 1 * * MON", AnnounceWeeklyRotations)
	c.Start()
}

// AnnounceWeeklyRotations announces the weekly rotation
// for each channel by fetching the user list from the database
// and posting a message to the respective Slack channel
func AnnounceWeeklyRotations() {
	slack := slack.New(config.SlackBotToken)
	rows, err := db.DB.Query(`SELECT DISTINCT channel_id FROM users`)
	if err != nil {
		log.Printf("Failed to fetch channel IDs: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			log.Printf("Failed to scan channel ID: %v", err)
			continue
		}
		announceRotationForChannel(channelID, slack)
	}
}

// announceRotationForChannel announces the rotation for a specific channel
func announceRotationForChannel(channelID string, slackClient *slack.Client) {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("DB transaction error for channel %s: %v", channelID, err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT user_id FROM users
		WHERE channel_id = ?
		ORDER BY position ASC
	`, channelID)
	if err != nil {
		log.Printf("Failed to fetch users for channel %s: %v", channelID, err)
		return
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			log.Printf("Scan error: %v", err)
			return
		}
		userIDs = append(userIDs, uid)
	}

	if len(userIDs) == 0 {
		log.Printf("No users found for channel %s", channelID)
		return
	}

	// Announce first person in rotation
	firstUser := userIDs[0]
	// we may need to get the userIds from slack?
	message := slack.MsgOptionText("🔁 This week's release train engineer: <@"+firstUser+">", false)
	if _, _, err := slackClient.PostMessage(channelID, message); err != nil {
		log.Printf("Failed to post to Slack channel %s: %v", channelID, err)
		return
	}

	// Rotate the list
	for i, uid := range append(userIDs[1:], firstUser) {
		_, err := tx.Exec(`
			UPDATE users SET position = ? WHERE user_id = ? AND channel_id = ?
		`, i, uid, channelID)
		if err != nil {
			log.Printf("Failed to rotate user %s: %v", uid, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Commit failed for channel %s: %v", channelID, err)
		return
	}
}
