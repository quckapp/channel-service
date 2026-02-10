package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/api"
	"github.com/quckapp/channel-service/internal/config"
	"github.com/quckapp/channel-service/internal/db"
	"github.com/quckapp/channel-service/internal/repository"
	"github.com/quckapp/channel-service/internal/service"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting channel service...")

	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	if cfg.Environment == "development" {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Initialize MySQL
	mysqlDB, err := db.NewMySQL(cfg.DatabaseURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MySQL")
	}
	defer mysqlDB.Close()
	logger.Info("Connected to MySQL database")

	// Run migrations
	if err := runMigrations(mysqlDB); err != nil {
		logger.WithError(err).Warn("Failed to run migrations")
	}

	// Initialize Redis
	redisClient, err := db.NewRedis(cfg.RedisURL)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect to Redis, continuing without cache")
		redisClient = nil
	} else {
		defer redisClient.Close()
		logger.Info("Connected to Redis")
	}

	// Initialize Kafka
	var kafkaProducer *db.KafkaProducer
	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" {
		kafkaProducer, err = db.NewKafkaProducer(cfg.KafkaBrokers)
		if err != nil {
			logger.WithError(err).Warn("Failed to connect to Kafka, continuing without events")
			kafkaProducer = nil
		} else {
			defer kafkaProducer.Close()
			logger.Info("Connected to Kafka")
		}
	}

	// Initialize repositories
	channelRepo := repository.NewChannelRepository(mysqlDB)
	memberRepo := repository.NewMemberRepository(mysqlDB)
	pinRepo := repository.NewPinRepository(mysqlDB)
	inviteRepo := repository.NewInviteRepository(mysqlDB)
	bookmarkRepo := repository.NewBookmarkRepository(mysqlDB)
	topicHistoryRepo := repository.NewTopicHistoryRepository(mysqlDB)
	permissionRepo := repository.NewPermissionRepository(mysqlDB)
	webhookRepo := repository.NewWebhookRepository(mysqlDB)
	reactionRepo := repository.NewReactionRepository(mysqlDB)
	moderationRepo := repository.NewModerationRepository(mysqlDB)
	announcementRepo := repository.NewAnnouncementRepository(mysqlDB)
	sectionRepo := repository.NewSectionRepository(mysqlDB)
	analyticsRepo := repository.NewAnalyticsRepository(mysqlDB)
	threadRepo := repository.NewThreadRepository(mysqlDB)
	settingsRepo := repository.NewSettingsRepository(mysqlDB)
	starredRepo := repository.NewStarredRepository(mysqlDB)
	readReceiptRepo := repository.NewReadReceiptRepository(mysqlDB)
	scheduledMessageRepo := repository.NewScheduledMessageRepository(mysqlDB)
	activityLogRepo := repository.NewActivityLogRepository(mysqlDB)
	templateRepo := repository.NewTemplateRepository(mysqlDB)
	voiceRepo := repository.NewVoiceRepository(mysqlDB)
	followerRepo := repository.NewFollowerRepository(mysqlDB)

	// Initialize service
	channelService := service.NewChannelService(
		channelRepo,
		memberRepo,
		pinRepo,
		inviteRepo,
		bookmarkRepo,
		topicHistoryRepo,
		permissionRepo,
		webhookRepo,
		reactionRepo,
		moderationRepo,
		announcementRepo,
		sectionRepo,
		analyticsRepo,
		threadRepo,
		settingsRepo,
		starredRepo,
		readReceiptRepo,
		scheduledMessageRepo,
		activityLogRepo,
		templateRepo,
		voiceRepo,
		followerRepo,
		redisClient,
		kafkaProducer,
		logger,
	)

	// Initialize router
	router := api.NewRouter(channelService, cfg, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.WithField("port", cfg.Port).Info("Channel service listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down channel service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Channel service stopped")
}

func runMigrations(db *sqlx.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS channels (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			type ENUM('public', 'private', 'dm', 'group_dm') DEFAULT 'public',
			description TEXT,
			topic VARCHAR(500),
			icon_url VARCHAR(500),
			is_archived BOOLEAN DEFAULT FALSE,
			created_by CHAR(36),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_channels_workspace (workspace_id),
			INDEX idx_channels_type (type),
			UNIQUE KEY unique_channel_name (workspace_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_members (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			role ENUM('owner', 'admin', 'member') DEFAULT 'member',
			notifications ENUM('all', 'mentions', 'none') DEFAULT 'all',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_read_at TIMESTAMP NULL,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_channel_member (channel_id, user_id),
			INDEX idx_channel_members_user (user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_pins (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			message_id CHAR(36) NOT NULL,
			pinned_by CHAR(36) NOT NULL,
			pinned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_pin (channel_id, message_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_invites (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			created_by CHAR(36) NOT NULL,
			code VARCHAR(20) NOT NULL UNIQUE,
			max_uses INT DEFAULT 0,
			use_count INT DEFAULT 0,
			expires_at TIMESTAMP NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_invite_code (code),
			INDEX idx_invite_channel (channel_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_bookmarks (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			title VARCHAR(255) NOT NULL,
			url VARCHAR(2000),
			entity_type VARCHAR(50),
			entity_id CHAR(36),
			position INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_bookmark_user (channel_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_topic_history (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			old_topic VARCHAR(500),
			new_topic VARCHAR(500),
			changed_by CHAR(36) NOT NULL,
			changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_topic_history_channel (channel_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_permissions (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			permission_type VARCHAR(100) NOT NULL,
			target_type ENUM('role', 'user') NOT NULL,
			target_id VARCHAR(100) NOT NULL,
			allow BOOLEAN DEFAULT FALSE,
			deny BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_perm (channel_id, permission_type, target_type, target_id),
			INDEX idx_perm_channel (channel_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_webhooks (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			url VARCHAR(2000) NOT NULL,
			avatar_url VARCHAR(500),
			events JSON,
			is_active BOOLEAN DEFAULT TRUE,
			created_by CHAR(36) NOT NULL,
			last_triggered_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_webhook_channel (channel_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_reactions (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			message_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			emoji VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_reaction (channel_id, message_id, user_id, emoji),
			INDEX idx_reaction_message (channel_id, message_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_bans (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			banned_by CHAR(36) NOT NULL,
			reason TEXT,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_ban (channel_id, user_id),
			INDEX idx_ban_channel (channel_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_mutes (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			muted_by CHAR(36) NOT NULL,
			reason TEXT,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_mute (channel_id, user_id),
			INDEX idx_mute_channel (channel_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_moderation_log (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			action VARCHAR(50) NOT NULL,
			actor_id CHAR(36) NOT NULL,
			reason TEXT,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_modlog_channel (channel_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_announcements (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			priority ENUM('low', 'normal', 'high', 'urgent') DEFAULT 'normal',
			author_id CHAR(36) NOT NULL,
			is_pinned BOOLEAN DEFAULT FALSE,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_announcement_channel (channel_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_sections (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			position INT DEFAULT 0,
			is_collapsed BOOLEAN DEFAULT FALSE,
			channel_ids JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_section_user (workspace_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_threads (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			message_id CHAR(36) NOT NULL,
			title VARCHAR(255),
			created_by CHAR(36) NOT NULL,
			is_locked BOOLEAN DEFAULT FALSE,
			is_resolved BOOLEAN DEFAULT FALSE,
			reply_count INT DEFAULT 0,
			last_reply_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_thread_channel (channel_id),
			INDEX idx_thread_message (message_id)
		)`,
		`CREATE TABLE IF NOT EXISTS thread_replies (
			id CHAR(36) PRIMARY KEY,
			thread_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			content TEXT NOT NULL,
			parent_id CHAR(36),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (thread_id) REFERENCES channel_threads(id) ON DELETE CASCADE,
			INDEX idx_reply_thread (thread_id)
		)`,
		`CREATE TABLE IF NOT EXISTS thread_followers (
			id CHAR(36) PRIMARY KEY,
			thread_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (thread_id) REFERENCES channel_threads(id) ON DELETE CASCADE,
			UNIQUE KEY unique_thread_follower (thread_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_settings (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL UNIQUE,
			slow_mode_interval INT DEFAULT 0,
			max_pins INT DEFAULT 50,
			max_bookmarks INT DEFAULT 100,
			allow_threads BOOLEAN DEFAULT TRUE,
			allow_reactions BOOLEAN DEFAULT TRUE,
			allow_invites BOOLEAN DEFAULT TRUE,
			auto_archive_days INT DEFAULT 0,
			default_notification VARCHAR(50) DEFAULT 'all',
			custom_emoji BOOLEAN DEFAULT FALSE,
			link_previews BOOLEAN DEFAULT TRUE,
			member_limit INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS starred_channels (
			id CHAR(36) PRIMARY KEY,
			user_id CHAR(36) NOT NULL,
			channel_id CHAR(36) NOT NULL,
			position INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_star (user_id, channel_id),
			INDEX idx_starred_user (user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_read_receipts (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			message_id CHAR(36) NOT NULL,
			read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_read_receipt (channel_id, user_id, message_id),
			INDEX idx_receipt_message (channel_id, message_id)
		)`,
		`CREATE TABLE IF NOT EXISTS scheduled_messages (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			content TEXT NOT NULL,
			scheduled_at TIMESTAMP NOT NULL,
			status ENUM('pending', 'sent', 'cancelled', 'failed') DEFAULT 'pending',
			sent_at TIMESTAMP NULL,
			thread_id CHAR(36),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_scheduled_channel (channel_id),
			INDEX idx_scheduled_user (user_id),
			INDEX idx_scheduled_status (status, scheduled_at)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_activity_log (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			action VARCHAR(100) NOT NULL,
			target_id CHAR(36),
			details TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_activity_channel (channel_id),
			INDEX idx_activity_user (user_id),
			INDEX idx_activity_action (action)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_templates (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			type VARCHAR(20) DEFAULT 'public',
			topic VARCHAR(500),
			settings JSON,
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_template_workspace (workspace_id)
		)`,
		`CREATE TABLE IF NOT EXISTS voice_channel_states (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			is_muted BOOLEAN DEFAULT FALSE,
			is_deafened BOOLEAN DEFAULT FALSE,
			is_screen_share BOOLEAN DEFAULT FALSE,
			is_video_on BOOLEAN DEFAULT FALSE,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			disconnected_at TIMESTAMP NULL,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			INDEX idx_voice_channel (channel_id),
			INDEX idx_voice_active (channel_id, disconnected_at)
		)`,
		`CREATE TABLE IF NOT EXISTS channel_followers (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			followed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_follower (channel_id, user_id),
			INDEX idx_follower_channel (channel_id),
			INDEX idx_follower_user (user_id)
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}
