package api

import (
	"net/http"
	"strconv"

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

// ── Channel CRUD ──

func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	userID := getUserID(c)
	var req models.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.service.CreateChannel(c.Request.Context(), userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ch)
}

func (h *ChannelHandler) GetChannel(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	resp, err := h.service.GetChannel(c.Request.Context(), id, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var req models.UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.service.UpdateChannel(c.Request.Context(), id, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ch)
}

func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	if err := h.service.DeleteChannel(c.Request.Context(), id, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ChannelHandler) ListChannels(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	channels, err := h.service.ListChannels(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

func (h *ChannelHandler) ListUserChannels(c *gin.Context) {
	userID := getUserID(c)
	workspaceID := c.Query("workspace_id")

	channels, err := h.service.ListUserChannels(c.Request.Context(), userID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

func (h *ChannelHandler) SearchChannels(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	query := c.Query("q")

	channels, err := h.service.SearchChannels(c.Request.Context(), workspaceID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

// ── Archive ──

func (h *ChannelHandler) ArchiveChannel(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	if err := h.service.ArchiveChannel(c.Request.Context(), id, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel archived"})
}

func (h *ChannelHandler) UnarchiveChannel(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	if err := h.service.UnarchiveChannel(c.Request.Context(), id, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel unarchived"})
}

// ── Members ──

func (h *ChannelHandler) AddMember(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.service.AddMember(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, member)
}

func (h *ChannelHandler) RemoveMember(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	memberUserID := c.Param("userId")

	if err := h.service.RemoveMember(c.Request.Context(), channelID, memberUserID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ChannelHandler) LeaveChannel(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	if err := h.service.LeaveChannel(c.Request.Context(), channelID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Left channel"})
}

func (h *ChannelHandler) GetMember(c *gin.Context) {
	channelID := c.Param("id")
	memberUserID := c.Param("userId")

	member, err := h.service.GetMember(c.Request.Context(), channelID, memberUserID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, member)
}

func (h *ChannelHandler) ListMembers(c *gin.Context) {
	channelID := c.Param("id")

	members, err := h.service.ListMembers(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (h *ChannelHandler) UpdateMemberRole(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	memberUserID := c.Param("userId")

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateMemberRole(c.Request.Context(), channelID, memberUserID, userID, req.Role); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}

func (h *ChannelHandler) UpdateNotifications(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.UpdateNotificationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateNotifications(c.Request.Context(), channelID, userID, req.Notifications); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notifications updated"})
}

func (h *ChannelHandler) UpdateLastRead(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	if err := h.service.UpdateLastRead(c.Request.Context(), channelID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Last read updated"})
}

// ── Pins ──

func (h *ChannelHandler) PinMessage(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.PinMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pin, err := h.service.PinMessage(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, pin)
}

func (h *ChannelHandler) UnpinMessage(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	messageID := c.Param("messageId")

	if err := h.service.UnpinMessage(c.Request.Context(), channelID, messageID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message unpinned"})
}

func (h *ChannelHandler) ListPins(c *gin.Context) {
	channelID := c.Param("id")

	pins, err := h.service.ListPins(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pins"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pins": pins})
}

// ── Typing ──

func (h *ChannelHandler) SetTyping(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	h.service.SetTyping(c.Request.Context(), channelID, userID)
	c.JSON(http.StatusOK, gin.H{"message": "Typing set"})
}

func (h *ChannelHandler) GetTyping(c *gin.Context) {
	channelID := c.Param("id")

	users, err := h.service.GetTyping(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get typing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"typing": users})
}

// ── Invites ──

func (h *ChannelHandler) CreateInvite(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invite, err := h.service.CreateInvite(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, invite)
}

func (h *ChannelHandler) ListInvites(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	invites, err := h.service.ListInvites(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"invites": invites})
}

func (h *ChannelHandler) DeleteInvite(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	inviteID := c.Param("inviteId")

	if err := h.service.DeleteInvite(c.Request.Context(), channelID, inviteID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ChannelHandler) JoinByCode(c *gin.Context) {
	userID := getUserID(c)
	code := c.Param("code")

	member, err := h.service.JoinByCode(c.Request.Context(), code, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, member)
}

// ── Bookmarks ──

func (h *ChannelHandler) CreateBookmark(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookmark, err := h.service.CreateBookmark(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, bookmark)
}

func (h *ChannelHandler) ListBookmarks(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	bookmarks, err := h.service.ListBookmarks(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookmarks": bookmarks})
}

func (h *ChannelHandler) UpdateBookmark(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	bookmarkID := c.Param("bookmarkId")

	var req models.UpdateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookmark, err := h.service.UpdateBookmark(c.Request.Context(), channelID, bookmarkID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, bookmark)
}

func (h *ChannelHandler) DeleteBookmark(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	bookmarkID := c.Param("bookmarkId")

	if err := h.service.DeleteBookmark(c.Request.Context(), channelID, bookmarkID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ── Topic History ──

func (h *ChannelHandler) GetTopicHistory(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	history, err := h.service.GetTopicHistory(c.Request.Context(), channelID, userID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"topic_history": history})
}

// ── Permissions ──

func (h *ChannelHandler) SetPermission(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.SetPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	perm, err := h.service.SetPermission(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, perm)
}

func (h *ChannelHandler) ListPermissions(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	perms, err := h.service.ListPermissions(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": perms})
}

func (h *ChannelHandler) DeletePermission(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	permissionID := c.Param("permissionId")

	if err := h.service.DeletePermission(c.Request.Context(), channelID, permissionID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ── Webhooks ──

func (h *ChannelHandler) CreateWebhook(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.service.CreateWebhook(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

func (h *ChannelHandler) ListWebhooks(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	webhooks, err := h.service.ListWebhooks(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"webhooks": webhooks})
}

func (h *ChannelHandler) UpdateWebhook(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	webhookID := c.Param("webhookId")

	var req models.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.service.UpdateWebhook(c.Request.Context(), channelID, webhookID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, webhook)
}

func (h *ChannelHandler) DeleteWebhook(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	webhookID := c.Param("webhookId")

	if err := h.service.DeleteWebhook(c.Request.Context(), channelID, webhookID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ChannelHandler) TestWebhook(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	webhookID := c.Param("webhookId")

	if err := h.service.TestWebhook(c.Request.Context(), channelID, webhookID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook test triggered"})
}

// ── Reactions ──

func (h *ChannelHandler) AddReaction(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.AddReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reaction, err := h.service.AddReaction(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, reaction)
}

func (h *ChannelHandler) RemoveReaction(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.RemoveReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.RemoveReaction(c.Request.Context(), channelID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}

func (h *ChannelHandler) ListReactions(c *gin.Context) {
	channelID := c.Param("id")
	messageID := c.Query("message_id")

	reactions, err := h.service.ListReactions(c.Request.Context(), channelID, messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list reactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reactions": reactions})
}

func (h *ChannelHandler) GetReactionSummary(c *gin.Context) {
	channelID := c.Param("id")
	messageID := c.Query("message_id")

	summary, err := h.service.GetReactionSummary(c.Request.Context(), channelID, messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reaction summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

// ── Moderation ──

func (h *ChannelHandler) BanMember(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	targetUserID := c.Param("userId")

	var req models.BanMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.BanMember(c.Request.Context(), channelID, targetUserID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member banned"})
}

func (h *ChannelHandler) UnbanMember(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	targetUserID := c.Param("userId")

	if err := h.service.UnbanMember(c.Request.Context(), channelID, targetUserID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member unbanned"})
}

func (h *ChannelHandler) MuteMember(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	targetUserID := c.Param("userId")

	var req models.MuteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.MuteMember(c.Request.Context(), channelID, targetUserID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member muted"})
}

func (h *ChannelHandler) UnmuteMember(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	targetUserID := c.Param("userId")

	if err := h.service.UnmuteMember(c.Request.Context(), channelID, targetUserID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member unmuted"})
}

func (h *ChannelHandler) GetModerationHistory(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	entries, err := h.service.GetModerationHistory(c.Request.Context(), channelID, userID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"moderation_history": entries})
}

// ── Announcements ──

func (h *ChannelHandler) CreateAnnouncement(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ann, err := h.service.CreateAnnouncement(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ann)
}

func (h *ChannelHandler) ListAnnouncements(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	anns, err := h.service.ListAnnouncements(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"announcements": anns})
}

func (h *ChannelHandler) UpdateAnnouncement(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	announcementID := c.Param("announcementId")

	var req models.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ann, err := h.service.UpdateAnnouncement(c.Request.Context(), channelID, announcementID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ann)
}

func (h *ChannelHandler) DeleteAnnouncement(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	announcementID := c.Param("announcementId")

	if err := h.service.DeleteAnnouncement(c.Request.Context(), channelID, announcementID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ChannelHandler) PinAnnouncement(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	announcementID := c.Param("announcementId")

	if err := h.service.PinAnnouncement(c.Request.Context(), channelID, announcementID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Announcement pin toggled"})
}

// ── Sections ──

func (h *ChannelHandler) CreateSection(c *gin.Context) {
	userID := getUserID(c)

	var req models.CreateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	section, err := h.service.CreateSection(c.Request.Context(), userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, section)
}

func (h *ChannelHandler) ListSections(c *gin.Context) {
	userID := getUserID(c)
	workspaceID := c.Query("workspace_id")

	sections, err := h.service.ListSections(c.Request.Context(), workspaceID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sections"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sections": sections})
}

func (h *ChannelHandler) UpdateSection(c *gin.Context) {
	userID := getUserID(c)
	sectionID := c.Param("sectionId")

	var req models.UpdateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	section, err := h.service.UpdateSection(c.Request.Context(), sectionID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, section)
}

func (h *ChannelHandler) DeleteSection(c *gin.Context) {
	userID := getUserID(c)
	sectionID := c.Param("sectionId")

	if err := h.service.DeleteSection(c.Request.Context(), sectionID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ── Analytics & Stats ──

func (h *ChannelHandler) GetChannelStats(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	stats, err := h.service.GetChannelStats(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *ChannelHandler) GetChannelAnalytics(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	activities, err := h.service.GetChannelAnalytics(c.Request.Context(), channelID, userID, days)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"analytics": activities})
}

// ── Transfer Ownership ──

func (h *ChannelHandler) TransferOwnership(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.TransferOwnershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.TransferOwnership(c.Request.Context(), channelID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ownership transferred"})
}

// ── Bulk Member Operations ──

func (h *ChannelHandler) BulkAddMembers(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.BulkAddMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	added, err := h.service.BulkAddMembers(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"added": added})
}

// ── Clone Channel ──

func (h *ChannelHandler) CloneChannel(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CloneChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.service.CloneChannel(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ch)
}

// ── Threads ──

func (h *ChannelHandler) CreateThread(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	thread, err := h.service.CreateThread(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, thread)
}

func (h *ChannelHandler) GetThread(c *gin.Context) {
	threadID := c.Param("threadId")

	thread, err := h.service.GetThread(c.Request.Context(), threadID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, thread)
}

func (h *ChannelHandler) ListThreads(c *gin.Context) {
	channelID := c.Param("id")
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	threads, err := h.service.ListThreads(c.Request.Context(), channelID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, threads)
}

func (h *ChannelHandler) DeleteThread(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	threadID := c.Param("threadId")

	if err := h.service.DeleteThread(c.Request.Context(), channelID, threadID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *ChannelHandler) CreateReply(c *gin.Context) {
	userID := getUserID(c)
	threadID := c.Param("threadId")

	var req models.CreateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reply, err := h.service.CreateReply(c.Request.Context(), threadID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, reply)
}

func (h *ChannelHandler) ListReplies(c *gin.Context) {
	threadID := c.Param("threadId")
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	replies, err := h.service.ListReplies(c.Request.Context(), threadID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, replies)
}

func (h *ChannelHandler) UpdateReply(c *gin.Context) {
	userID := getUserID(c)
	replyID := c.Param("replyId")

	var req models.UpdateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reply, err := h.service.UpdateReply(c.Request.Context(), replyID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, reply)
}

func (h *ChannelHandler) DeleteReply(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	replyID := c.Param("replyId")

	if err := h.service.DeleteReply(c.Request.Context(), channelID, replyID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *ChannelHandler) FollowThread(c *gin.Context) {
	userID := getUserID(c)
	threadID := c.Param("threadId")

	if err := h.service.FollowThread(c.Request.Context(), threadID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"following": true})
}

func (h *ChannelHandler) UnfollowThread(c *gin.Context) {
	userID := getUserID(c)
	threadID := c.Param("threadId")

	if err := h.service.UnfollowThread(c.Request.Context(), threadID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"following": false})
}

// ── Channel Settings ──

func (h *ChannelHandler) GetSettings(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	settings, err := h.service.GetSettings(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *ChannelHandler) UpdateSettings(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.UpdateSettings(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, settings)
}

// ── Starred Channels ──

func (h *ChannelHandler) StarChannel(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	star, err := h.service.StarChannel(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, star)
}

func (h *ChannelHandler) UnstarChannel(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	if err := h.service.UnstarChannel(c.Request.Context(), channelID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"unstarred": true})
}

func (h *ChannelHandler) ListStarredChannels(c *gin.Context) {
	userID := getUserID(c)

	starred, err := h.service.ListStarredChannels(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, starred)
}

// ── Read Receipts ──

func (h *ChannelHandler) MarkReadReceipt(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.MarkRead(c.Request.Context(), channelID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"marked": true})
}

func (h *ChannelHandler) GetReadReceipts(c *gin.Context) {
	channelID := c.Param("id")
	messageID := c.Query("message_id")

	receipts, err := h.service.GetReadReceipts(c.Request.Context(), channelID, messageID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, receipts)
}

func (h *ChannelHandler) GetReadCount(c *gin.Context) {
	channelID := c.Param("id")
	messageID := c.Query("message_id")

	count, err := h.service.GetReadCount(c.Request.Context(), channelID, messageID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// ── Scheduled Messages ──

func (h *ChannelHandler) CreateScheduledMessage(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.CreateScheduledMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.service.CreateScheduledMessage(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, msg)
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

func (h *ChannelHandler) ListScheduledMessages(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	messages, err := h.service.ListScheduledMessages(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *ChannelHandler) ListMyScheduledMessages(c *gin.Context) {
	userID := getUserID(c)

	messages, err := h.service.ListMyScheduledMessages(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, messages)
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

func (h *ChannelHandler) DeleteScheduledMessage(c *gin.Context) {
	userID := getUserID(c)
	messageID := c.Param("messageId")

	if err := h.service.DeleteScheduledMessage(c.Request.Context(), messageID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// ── Activity Log ──

func (h *ChannelHandler) GetActivityLog(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, err := h.service.GetActivityLog(c.Request.Context(), channelID, userID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *ChannelHandler) GetUserActivityLog(c *gin.Context) {
	userID := getUserID(c)
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, err := h.service.GetUserActivityLog(c.Request.Context(), userID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}

// ── Channel Templates ──

func (h *ChannelHandler) CreateTemplate(c *gin.Context) {
	userID := getUserID(c)

	var req models.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.service.CreateTemplate(c.Request.Context(), userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, tmpl)
}

func (h *ChannelHandler) GetTemplate(c *gin.Context) {
	templateID := c.Param("templateId")

	tmpl, err := h.service.GetTemplate(c.Request.Context(), templateID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

func (h *ChannelHandler) ListTemplates(c *gin.Context) {
	workspaceID := c.Query("workspace_id")

	templates, err := h.service.ListTemplates(c.Request.Context(), workspaceID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, templates)
}

func (h *ChannelHandler) UpdateTemplate(c *gin.Context) {
	userID := getUserID(c)
	templateID := c.Param("templateId")

	var req models.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.service.UpdateTemplate(c.Request.Context(), templateID, userID, &req)
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

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *ChannelHandler) ApplyTemplate(c *gin.Context) {
	userID := getUserID(c)
	templateID := c.Param("templateId")

	var req models.ApplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.service.ApplyTemplate(c.Request.Context(), templateID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ch)
}

// ── Voice Channels ──

func (h *ChannelHandler) JoinVoice(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	state, err := h.service.JoinVoice(c.Request.Context(), channelID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, state)
}

func (h *ChannelHandler) LeaveVoice(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	if err := h.service.LeaveVoice(c.Request.Context(), channelID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"left": true})
}

func (h *ChannelHandler) UpdateVoiceState(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.UpdateVoiceStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateVoiceState(c.Request.Context(), channelID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": true})
}

func (h *ChannelHandler) ListVoiceParticipants(c *gin.Context) {
	channelID := c.Param("id")

	participants, err := h.service.ListVoiceParticipants(c.Request.Context(), channelID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, participants)
}

// ── Channel Followers ──

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

	c.JSON(http.StatusOK, gin.H{"unfollowed": true})
}

func (h *ChannelHandler) ListChannelFollowers(c *gin.Context) {
	channelID := c.Param("id")
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	followers, err := h.service.ListChannelFollowers(c.Request.Context(), channelID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, followers)
}

func (h *ChannelHandler) ListFollowedChannels(c *gin.Context) {
	userID := getUserID(c)

	followed, err := h.service.ListFollowedChannels(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, followed)
}

func (h *ChannelHandler) CountChannelFollowers(c *gin.Context) {
	channelID := c.Param("id")

	count, err := h.service.CountChannelFollowers(c.Request.Context(), channelID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// ── Bulk Operations Extended ──

func (h *ChannelHandler) BulkDeleteMembers(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.BulkDeleteMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.BulkDeleteMembers(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ChannelHandler) BulkUpdateRoles(c *gin.Context) {
	userID := getUserID(c)
	channelID := c.Param("id")

	var req models.BulkUpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.BulkUpdateRoles(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ── Helpers ──

func getUserID(c *gin.Context) string {
	userIDStr, _ := c.Get("user_id")
	return userIDStr.(string)
}

func handleError(c *gin.Context, err error) {
	switch err {
	case service.ErrChannelNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
	case service.ErrNotAuthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
	case service.ErrNotMember:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this channel"})
	case service.ErrAlreadyMember:
		c.JSON(http.StatusConflict, gin.H{"error": "Already a member"})
	case service.ErrAlreadyPinned:
		c.JSON(http.StatusConflict, gin.H{"error": "Message already pinned"})
	case service.ErrPinNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Pin not found"})
	case service.ErrChannelArchived:
		c.JSON(http.StatusConflict, gin.H{"error": "Channel is archived"})
	case service.ErrCannotLeaveOwner:
		c.JSON(http.StatusConflict, gin.H{"error": "Owner cannot leave channel"})
	case service.ErrInviteNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found"})
	case service.ErrInviteExpired:
		c.JSON(http.StatusGone, gin.H{"error": "Invite has expired"})
	case service.ErrInviteMaxUses:
		c.JSON(http.StatusConflict, gin.H{"error": "Invite has reached max uses"})
	case service.ErrBookmarkNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Bookmark not found"})
	case service.ErrBookmarkLimitReached:
		c.JSON(http.StatusConflict, gin.H{"error": "Bookmark limit reached"})
	case service.ErrPermissionNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
	case service.ErrWebhookNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
	case service.ErrReactionExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Reaction already exists"})
	case service.ErrReactionNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Reaction not found"})
	case service.ErrUserBanned:
		c.JSON(http.StatusForbidden, gin.H{"error": "User is banned from this channel"})
	case service.ErrUserMuted:
		c.JSON(http.StatusForbidden, gin.H{"error": "User is muted in this channel"})
	case service.ErrAnnouncementNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
	case service.ErrSectionNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Section not found"})
	case service.ErrThreadNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
	case service.ErrReplyNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Reply not found"})
	case service.ErrSettingsNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel settings not found"})
	case service.ErrScheduledMessageNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Scheduled message not found"})
	case service.ErrTemplateNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
	case service.ErrAlreadyStarred:
		c.JSON(http.StatusConflict, gin.H{"error": "Channel already starred"})
	case service.ErrNotStarred:
		c.JSON(http.StatusConflict, gin.H{"error": "Channel not starred"})
	case service.ErrNotInVoice:
		c.JSON(http.StatusConflict, gin.H{"error": "Not in voice channel"})
	case service.ErrAlreadyInVoice:
		c.JSON(http.StatusConflict, gin.H{"error": "Already in voice channel"})
	case service.ErrAlreadyFollowing:
		c.JSON(http.StatusConflict, gin.H{"error": "Already following channel"})
	case service.ErrNotFollowing:
		c.JSON(http.StatusConflict, gin.H{"error": "Not following channel"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
