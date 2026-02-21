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
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting channel service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	if cfg.Environment == "development" {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.WithFields(logrus.Fields{
		"port":        cfg.Port,
		"environment": cfg.Environment,
	}).Info("Configuration loaded")

	// Initialize MySQL database
	mysqlDB, err := db.NewMySQL(cfg.DatabaseURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MySQL")
	}
	defer mysqlDB.Close()
	logger.Info("Connected to MySQL database")

	// Run database migrations
	if err := runMigrations(mysqlDB); err != nil {
		logger.WithError(err).Warn("Failed to run migrations")
	}

	// Initialize Redis
	_, err = db.NewRedis(cfg.RedisURL)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect to Redis, continuing without cache")
	} else {
		logger.Info("Connected to Redis")
	}

	// Initialize Kafka producer
	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" {
		kafkaProducer, err := db.NewKafkaProducer(cfg.KafkaBrokers)
		if err != nil {
			logger.WithError(err).Warn("Failed to connect to Kafka, continuing without events")
		} else {
			defer kafkaProducer.Close()
			logger.WithField("brokers", cfg.KafkaBrokers).Info("Connected to Kafka")
		}
	}

	// Initialize repositories
	pollRepo := repository.NewPollRepository(mysqlDB)
	scheduledMessageRepo := repository.NewScheduledMessageRepository(mysqlDB)
	channelLinkRepo := repository.NewChannelLinkRepository(mysqlDB)
	tabRepo := repository.NewTabRepository(mysqlDB)
	followerRepo := repository.NewFollowerRepository(mysqlDB)
	templateRepo := repository.NewTemplateRepository(mysqlDB)
	logger.Info("Repositories initialized")

	// Initialize service
	channelService := service.NewChannelService(
		pollRepo,
		scheduledMessageRepo,
		channelLinkRepo,
		tabRepo,
		followerRepo,
		templateRepo,
		logger,
	)
	logger.Info("Service layer initialized")

	// Initialize router
	router := api.NewRouter(channelService, cfg, logger)
	logger.Info("HTTP router initialized")

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithField("port", cfg.Port).Info("Channel service listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down channel service...")

	// Graceful shutdown with timeout
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
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
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
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS channel_pins (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			message_id CHAR(36) NOT NULL,
			pinned_by CHAR(36) NOT NULL,
			pinned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			UNIQUE KEY unique_pin (channel_id, message_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS channel_polls (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			created_by CHAR(36) NOT NULL,
			question VARCHAR(500) NOT NULL,
			is_anonymous BOOLEAN DEFAULT FALSE,
			multi_choice BOOLEAN DEFAULT FALSE,
			is_closed BOOLEAN DEFAULT FALSE,
			closed_at TIMESTAMP NULL,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_channel_polls_channel (channel_id),
			INDEX idx_channel_polls_created_by (created_by),
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS poll_options (
			id CHAR(36) PRIMARY KEY,
			poll_id CHAR(36) NOT NULL,
			text VARCHAR(200) NOT NULL,
			position INT DEFAULT 0,
			INDEX idx_poll_options_poll (poll_id),
			FOREIGN KEY (poll_id) REFERENCES channel_polls(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS poll_votes (
			id CHAR(36) PRIMARY KEY,
			poll_id CHAR(36) NOT NULL,
			option_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			voted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY unique_poll_user_option (poll_id, user_id, option_id),
			INDEX idx_poll_votes_poll (poll_id),
			INDEX idx_poll_votes_user (user_id),
			FOREIGN KEY (poll_id) REFERENCES channel_polls(id) ON DELETE CASCADE,
			FOREIGN KEY (option_id) REFERENCES poll_options(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS scheduled_messages (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			content TEXT NOT NULL,
			scheduled_at TIMESTAMP NOT NULL,
			sent_at TIMESTAMP NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_scheduled_messages_channel (channel_id),
			INDEX idx_scheduled_messages_user (user_id),
			INDEX idx_scheduled_messages_status (status),
			INDEX idx_scheduled_messages_scheduled_at (scheduled_at),
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS channel_links (
			id CHAR(36) PRIMARY KEY,
			source_channel_id CHAR(36) NOT NULL,
			target_channel_id CHAR(36) NOT NULL,
			created_by CHAR(36) NOT NULL,
			link_type VARCHAR(20) NOT NULL DEFAULT 'mirror',
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_channel_links_source (source_channel_id),
			INDEX idx_channel_links_target (target_channel_id),
			INDEX idx_channel_links_active (is_active),
			FOREIGN KEY (source_channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			FOREIGN KEY (target_channel_id) REFERENCES channels(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS channel_tabs (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			name VARCHAR(50) NOT NULL,
			tab_type VARCHAR(20) NOT NULL,
			config TEXT,
			position INT DEFAULT 0,
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_channel_tabs_channel (channel_id),
			INDEX idx_channel_tabs_position (position),
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS channel_followers (
			id CHAR(36) PRIMARY KEY,
			channel_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY unique_channel_follower (channel_id, user_id),
			INDEX idx_channel_followers_channel (channel_id),
			INDEX idx_channel_followers_user (user_id),
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS channel_templates (
			id CHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			created_by CHAR(36) NOT NULL,
			channel_type VARCHAR(20) DEFAULT 'public',
			default_topic VARCHAR(500),
			default_tabs TEXT,
			default_settings TEXT,
			is_public BOOLEAN DEFAULT FALSE,
			use_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_channel_templates_created_by (created_by),
			INDEX idx_channel_templates_is_public (is_public)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}
