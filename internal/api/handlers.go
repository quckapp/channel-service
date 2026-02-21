package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quckapp/channel-service/internal/models"
	"github.com/quckapp/channel-service/internal/service"
	"github.com/sirupsen/logrus"
)

type ChannelHandler struct {
	service *service.ChannelService
	logger  *logrus.Logger
}

func NewChannelHandler(svc *service.ChannelService, logger *logrus.Logger) *ChannelHandler {
	return &ChannelHandler{service: svc, logger: logger}
}

// ── Polls ──

func (h *ChannelHandler) CreatePoll(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	poll, err := h.service.CreatePoll(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, poll)
}

func (h *ChannelHandler) ListPolls(c *gin.Context) {
	channelID := c.Param("id")

	polls, err := h.service.ListPolls(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list polls"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"polls": polls})
}

func (h *ChannelHandler) GetPoll(c *gin.Context) {
	pollID := c.Param("pollId")

	poll, err := h.service.GetPoll(c.Request.Context(), pollID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, poll)
}

func (h *ChannelHandler) VotePoll(c *gin.Context) {
	userID := getUserID(c)
	pollID := c.Param("pollId")

	var req models.VotePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.VotePoll(c.Request.Context(), pollID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded"})
}

func (h *ChannelHandler) ClosePoll(c *gin.Context) {
	userID := getUserID(c)
	pollID := c.Param("pollId")

	if err := h.service.ClosePoll(c.Request.Context(), pollID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll closed"})
}

func (h *ChannelHandler) GetPollResults(c *gin.Context) {
	pollID := c.Param("pollId")

	results, err := h.service.GetPollResults(c.Request.Context(), pollID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, results)
}

// ── Scheduled Messages ──

func (h *ChannelHandler) ScheduleMessage(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateScheduledMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.service.ScheduleMessage(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *ChannelHandler) ListScheduledMessages(c *gin.Context) {
	channelID := c.Param("id")

	msgs, err := h.service.ListScheduledMessages(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list scheduled messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scheduled_messages": msgs})
}

func (h *ChannelHandler) GetScheduledMessage(c *gin.Context) {
	messageID := c.Param("messageId")

	msg, err := h.service.GetScheduledMessage(c.Request.Context(), messageID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, msg)
}

func (h *ChannelHandler) UpdateScheduledMessage(c *gin.Context) {
	userID := getUserID(c)
	messageID := c.Param("messageId")

	var req models.UpdateScheduledMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.service.UpdateScheduledMessage(c.Request.Context(), messageID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, msg)
}

func (h *ChannelHandler) CancelScheduledMessage(c *gin.Context) {
	userID := getUserID(c)
	messageID := c.Param("messageId")

	if err := h.service.CancelScheduledMessage(c.Request.Context(), messageID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ── Channel Links ──

func (h *ChannelHandler) CreateChannelLink(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateChannelLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.service.CreateChannelLink(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, link)
}

func (h *ChannelHandler) ListChannelLinks(c *gin.Context) {
	channelID := c.Param("id")

	links, err := h.service.ListChannelLinks(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list channel links"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"links": links})
}

func (h *ChannelHandler) GetChannelLink(c *gin.Context) {
	linkID := c.Param("linkId")

	link, err := h.service.GetChannelLink(c.Request.Context(), linkID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, link)
}

func (h *ChannelHandler) DeleteChannelLink(c *gin.Context) {
	userID := getUserID(c)
	linkID := c.Param("linkId")

	if err := h.service.DeleteChannelLink(c.Request.Context(), linkID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ── Tabs ──

func (h *ChannelHandler) AddTab(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateTabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tab, err := h.service.AddTab(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, tab)
}

func (h *ChannelHandler) ListTabs(c *gin.Context) {
	channelID := c.Param("id")

	tabs, err := h.service.ListTabs(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tabs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tabs": tabs})
}

func (h *ChannelHandler) UpdateTab(c *gin.Context) {
	userID := getUserID(c)
	tabID := c.Param("tabId")

	var req models.UpdateTabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tab, err := h.service.UpdateTab(c.Request.Context(), tabID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tab)
}

func (h *ChannelHandler) RemoveTab(c *gin.Context) {
	userID := getUserID(c)
	tabID := c.Param("tabId")

	if err := h.service.RemoveTab(c.Request.Context(), tabID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ChannelHandler) ReorderTabs(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.ReorderTabsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReorderTabs(c.Request.Context(), channelID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tabs reordered"})
}

// ── Followers ──

func (h *ChannelHandler) FollowChannel(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	follower, err := h.service.FollowChannel(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, follower)
}

func (h *ChannelHandler) UnfollowChannel(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	if err := h.service.UnfollowChannel(c.Request.Context(), channelID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ChannelHandler) ListFollowers(c *gin.Context) {
	channelID := c.Param("id")

	followers, err := h.service.ListFollowers(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list followers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"followers": followers})
}

func (h *ChannelHandler) CheckFollowing(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	isFollowing, err := h.service.CheckFollowing(c.Request.Context(), channelID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check following status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_following": isFollowing})
}

// ── Templates ──

func (h *ChannelHandler) CreateTemplate(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.service.CreateTemplate(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, tmpl)
}

func (h *ChannelHandler) ListTemplates(c *gin.Context) {
	templates, err := h.service.ListTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *ChannelHandler) ApplyTemplate(c *gin.Context) {
	userID := getUserID(c)
	templateID := c.Param("templateId")

	var req models.ApplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.service.ApplyTemplate(c.Request.Context(), templateID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

func (h *ChannelHandler) DeleteTemplate(c *gin.Context) {
	userID := getUserID(c)
	templateID := c.Param("templateId")

	if err := h.service.DeleteTemplate(c.Request.Context(), templateID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ── Helpers ──

func getUserID(c *gin.Context) string {
	userIDStr, _ := c.Get("user_id")
	return userIDStr.(string)
}

func handleError(c *gin.Context, err error) {
	switch err {
	case service.ErrPollNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
	case service.ErrAlreadyVoted:
		c.JSON(http.StatusConflict, gin.H{"error": "Already voted on this poll"})
	case service.ErrPollClosed:
		c.JSON(http.StatusConflict, gin.H{"error": "Poll is closed"})
	case service.ErrScheduledMessageNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Scheduled message not found"})
	case service.ErrScheduledTimeInPast:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Scheduled time must be in the future"})
	case service.ErrChannelLinkNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel link not found"})
	case service.ErrTabNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Tab not found"})
	case service.ErrAlreadyFollowing:
		c.JSON(http.StatusConflict, gin.H{"error": "Already following this channel"})
	case service.ErrNotFollowing:
		c.JSON(http.StatusNotFound, gin.H{"error": "Not following this channel"})
	case service.ErrChannelTemplateNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel template not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
