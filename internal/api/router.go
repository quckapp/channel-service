package api

import (
	"github.com/gin-gonic/gin"
	"github.com/quckapp/channel-service/internal/config"
	"github.com/quckapp/channel-service/internal/middleware"
	"github.com/quckapp/channel-service/internal/service"
	"github.com/sirupsen/logrus"
)

func NewRouter(
	channelService *service.ChannelService,
	cfg *config.Config,
	logger *logrus.Logger,
) *gin.Engine {
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
			// Polls
			channels.POST("/:id/polls", handler.CreatePoll)
			channels.GET("/:id/polls", handler.ListPolls)
			channels.GET("/:id/polls/:pollId", handler.GetPoll)
			channels.POST("/:id/polls/:pollId/vote", handler.VotePoll)
			channels.POST("/:id/polls/:pollId/close", handler.ClosePoll)
			channels.GET("/:id/polls/:pollId/results", handler.GetPollResults)

			// Scheduled Messages
			channels.POST("/:id/scheduled-messages", handler.ScheduleMessage)
			channels.GET("/:id/scheduled-messages", handler.ListScheduledMessages)
			channels.GET("/:id/scheduled-messages/:messageId", handler.GetScheduledMessage)
			channels.PUT("/:id/scheduled-messages/:messageId", handler.UpdateScheduledMessage)
			channels.DELETE("/:id/scheduled-messages/:messageId", handler.CancelScheduledMessage)

			// Channel Links
			channels.POST("/:id/links", handler.CreateChannelLink)
			channels.GET("/:id/links", handler.ListChannelLinks)
			channels.GET("/:id/links/:linkId", handler.GetChannelLink)
			channels.DELETE("/:id/links/:linkId", handler.DeleteChannelLink)

			// Tabs
			channels.POST("/:id/tabs", handler.AddTab)
			channels.GET("/:id/tabs", handler.ListTabs)
			channels.PUT("/:id/tabs/:tabId", handler.UpdateTab)
			channels.DELETE("/:id/tabs/:tabId", handler.RemoveTab)
			channels.POST("/:id/tabs/reorder", handler.ReorderTabs)

			// Followers
			channels.POST("/:id/follow", handler.FollowChannel)
			channels.DELETE("/:id/follow", handler.UnfollowChannel)
			channels.GET("/:id/followers", handler.ListFollowers)
			channels.GET("/:id/followers/check", handler.CheckFollowing)

			// Templates (channel-scoped)
			channels.POST("/:id/template", handler.CreateTemplate)
		}

		// Templates (standalone routes outside /:id)
		api.GET("/templates", middleware.Auth(cfg.JWTSecret), handler.ListTemplates)
		api.POST("/templates/:templateId/apply", middleware.Auth(cfg.JWTSecret), handler.ApplyTemplate)
		api.DELETE("/templates/:templateId", middleware.Auth(cfg.JWTSecret), handler.DeleteTemplate)
	}

	return r
}
