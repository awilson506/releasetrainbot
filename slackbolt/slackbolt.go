package slackbolt

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SlackCommandHandler map for slack command handlers
type SlackCommandHandler struct {
	Commands map[string]func(c *gin.Context)
}

// NewSlackCommandHandler initializes a new SlackCommandHandler
func NewSlackCommandHandler() *SlackCommandHandler {
	return &SlackCommandHandler{
		Commands: make(map[string]func(c *gin.Context)),
	}
}

// RegisterCommand allows you to register slash command handlers
func (h *SlackCommandHandler) RegisterCommand(command string, handler func(c *gin.Context)) {
	h.Commands[command] = handler
}

// Handle processes incoming Slack slash commands
func (h *SlackCommandHandler) Handle(c *gin.Context) {
	command := c.PostForm("command")
	if handler, found := h.Commands[command]; found {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "Unknown command")
	}
}
