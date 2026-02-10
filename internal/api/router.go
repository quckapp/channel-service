package api

import (
	"github.com/gin-gonic/gin"
	"github.com/quckapp/channel-service/internal/config"
	"github.com/quckapp/channel-service/internal/middleware"
	"github.com/quckapp/channel-service/internal/service"
	"github.com/sirupsen/logrus"
)

func NewRouter(channelService *service.ChannelService, cfg *config.Config, logger *logrus.Logger) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "channel-service"})
	})

	api := r.Group("/api/v1")
	{
		handler := NewChannelHandler(channelService, logger)

		channels := api.Group("/channels")
		channels.Use(middleware.Auth(cfg.JWTSecret))
		{
			// Channel CRUD
			channels.POST("", handler.CreateChannel)
			channels.GET("", handler.ListChannels)
			channels.GET("/me", handler.ListUserChannels)
			channels.GET("/search", handler.SearchChannels)
			channels.GET("/:id", handler.GetChannel)
			channels.PUT("/:id", handler.UpdateChannel)
			channels.DELETE("/:id", handler.DeleteChannel)

			// Archive
			channels.POST("/:id/archive", handler.ArchiveChannel)
			channels.POST("/:id/unarchive", handler.UnarchiveChannel)

			// Members
			channels.GET("/:id/members", handler.ListMembers)
			channels.GET("/:id/members/:userId", handler.GetMember)
			channels.POST("/:id/members", handler.AddMember)
			channels.DELETE("/:id/members/:userId", handler.RemoveMember)
			channels.PUT("/:id/members/:userId/role", handler.UpdateMemberRole)
			channels.POST("/:id/leave", handler.LeaveChannel)

			// Notifications & Read Status
			channels.PUT("/:id/notifications", handler.UpdateNotifications)
			channels.POST("/:id/read", handler.UpdateLastRead)

			// Pins
			channels.POST("/:id/pins", handler.PinMessage)
			channels.GET("/:id/pins", handler.ListPins)
			channels.DELETE("/:id/pins/:messageId", handler.UnpinMessage)

			// Typing
			channels.POST("/:id/typing", handler.SetTyping)
			channels.GET("/:id/typing", handler.GetTyping)

			// Invites
			channels.POST("/:id/invites", handler.CreateInvite)
			channels.GET("/:id/invites", handler.ListInvites)
			channels.DELETE("/:id/invites/:inviteId", handler.DeleteInvite)

			// Bookmarks
			channels.POST("/:id/bookmarks", handler.CreateBookmark)
			channels.GET("/:id/bookmarks", handler.ListBookmarks)
			channels.PUT("/:id/bookmarks/:bookmarkId", handler.UpdateBookmark)
			channels.DELETE("/:id/bookmarks/:bookmarkId", handler.DeleteBookmark)

			// Topic History
			channels.GET("/:id/topic-history", handler.GetTopicHistory)

			// Permissions
			channels.POST("/:id/permissions", handler.SetPermission)
			channels.GET("/:id/permissions", handler.ListPermissions)
			channels.DELETE("/:id/permissions/:permissionId", handler.DeletePermission)

			// Webhooks
			channels.POST("/:id/webhooks", handler.CreateWebhook)
			channels.GET("/:id/webhooks", handler.ListWebhooks)
			channels.PUT("/:id/webhooks/:webhookId", handler.UpdateWebhook)
			channels.DELETE("/:id/webhooks/:webhookId", handler.DeleteWebhook)
			channels.POST("/:id/webhooks/:webhookId/test", handler.TestWebhook)

			// Reactions
			channels.POST("/:id/reactions", handler.AddReaction)
			channels.DELETE("/:id/reactions", handler.RemoveReaction)
			channels.GET("/:id/reactions", handler.ListReactions)
			channels.GET("/:id/reactions/summary", handler.GetReactionSummary)

			// Moderation
			channels.POST("/:id/members/:userId/ban", handler.BanMember)
			channels.DELETE("/:id/members/:userId/ban", handler.UnbanMember)
			channels.POST("/:id/members/:userId/mute", handler.MuteMember)
			channels.DELETE("/:id/members/:userId/mute", handler.UnmuteMember)
			channels.GET("/:id/moderation", handler.GetModerationHistory)

			// Announcements
			channels.POST("/:id/announcements", handler.CreateAnnouncement)
			channels.GET("/:id/announcements", handler.ListAnnouncements)
			channels.PUT("/:id/announcements/:announcementId", handler.UpdateAnnouncement)
			channels.DELETE("/:id/announcements/:announcementId", handler.DeleteAnnouncement)
			channels.PUT("/:id/announcements/:announcementId/pin", handler.PinAnnouncement)

			// Analytics & Stats
			channels.GET("/:id/stats", handler.GetChannelStats)
			channels.GET("/:id/analytics", handler.GetChannelAnalytics)

			// Transfer Ownership
			channels.POST("/:id/transfer-ownership", handler.TransferOwnership)

			// Bulk Member Operations
			channels.POST("/:id/members/bulk", handler.BulkAddMembers)
			channels.DELETE("/:id/members/bulk", handler.BulkDeleteMembers)
			channels.PUT("/:id/members/bulk-role", handler.BulkUpdateRoles)

			// Clone Channel
			channels.POST("/:id/clone", handler.CloneChannel)

			// Threads
			channels.POST("/:id/threads", handler.CreateThread)
			channels.GET("/:id/threads", handler.ListThreads)
			channels.GET("/:id/threads/:threadId", handler.GetThread)
			channels.DELETE("/:id/threads/:threadId", handler.DeleteThread)
			channels.POST("/:id/threads/:threadId/replies", handler.CreateReply)
			channels.GET("/:id/threads/:threadId/replies", handler.ListReplies)
			channels.PUT("/:id/threads/:threadId/replies/:replyId", handler.UpdateReply)
			channels.DELETE("/:id/threads/:threadId/replies/:replyId", handler.DeleteReply)
			channels.POST("/:id/threads/:threadId/follow", handler.FollowThread)
			channels.DELETE("/:id/threads/:threadId/follow", handler.UnfollowThread)

			// Channel Settings
			channels.GET("/:id/settings", handler.GetSettings)
			channels.PUT("/:id/settings", handler.UpdateSettings)

			// Starred Channels
			channels.POST("/:id/star", handler.StarChannel)
			channels.DELETE("/:id/star", handler.UnstarChannel)

			// Read Receipts
			channels.POST("/:id/read-receipts", handler.MarkReadReceipt)
			channels.GET("/:id/read-receipts", handler.GetReadReceipts)
			channels.GET("/:id/read-receipts/count", handler.GetReadCount)

			// Scheduled Messages
			channels.POST("/:id/scheduled-messages", handler.CreateScheduledMessage)
			channels.GET("/:id/scheduled-messages", handler.ListScheduledMessages)
			channels.GET("/:id/scheduled-messages/:messageId", handler.GetScheduledMessage)
			channels.PUT("/:id/scheduled-messages/:messageId", handler.UpdateScheduledMessage)
			channels.DELETE("/:id/scheduled-messages/:messageId", handler.DeleteScheduledMessage)

			// Activity Log
			channels.GET("/:id/activity-log", handler.GetActivityLog)

			// Voice Channels
			channels.POST("/:id/voice/join", handler.JoinVoice)
			channels.POST("/:id/voice/leave", handler.LeaveVoice)
			channels.PUT("/:id/voice/state", handler.UpdateVoiceState)
			channels.GET("/:id/voice/participants", handler.ListVoiceParticipants)

			// Channel Followers
			channels.POST("/:id/follow", handler.FollowChannel)
			channels.DELETE("/:id/follow", handler.UnfollowChannel)
			channels.GET("/:id/followers", handler.ListChannelFollowers)
			channels.GET("/:id/followers/count", handler.CountChannelFollowers)
		}

		// Join by invite code (standalone, not scoped to /:id)
		api.POST("/channels/join/:code", middleware.Auth(cfg.JWTSecret), handler.JoinByCode)

		// Sections (user-scoped, not channel-scoped)
		sections := api.Group("/sections")
		sections.Use(middleware.Auth(cfg.JWTSecret))
		{
			sections.POST("", handler.CreateSection)
			sections.GET("", handler.ListSections)
			sections.PUT("/:sectionId", handler.UpdateSection)
			sections.DELETE("/:sectionId", handler.DeleteSection)
		}

		// Templates (workspace-scoped)
		templates := api.Group("/templates")
		templates.Use(middleware.Auth(cfg.JWTSecret))
		{
			templates.POST("", handler.CreateTemplate)
			templates.GET("", handler.ListTemplates)
			templates.GET("/:templateId", handler.GetTemplate)
			templates.PUT("/:templateId", handler.UpdateTemplate)
			templates.DELETE("/:templateId", handler.DeleteTemplate)
			templates.POST("/:templateId/apply", handler.ApplyTemplate)
		}

		// User-scoped convenience routes
		me := api.Group("/me")
		me.Use(middleware.Auth(cfg.JWTSecret))
		{
			me.GET("/starred-channels", handler.ListStarredChannels)
			me.GET("/followed-channels", handler.ListFollowedChannels)
			me.GET("/scheduled-messages", handler.ListMyScheduledMessages)
			me.GET("/activity-log", handler.GetUserActivityLog)
		}
	}

	return r
}
