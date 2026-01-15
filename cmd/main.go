package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	_ = godotenv.Load()

	db, err := sqlx.Connect("mysql", getEnv("DATABASE_URL", "root:password@tcp(localhost:3306)/quckapp_channels?parseTime=true"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "channel-service"})
	})

	api := r.Group("/api/v1")
	{
		channels := api.Group("/channels")
		{
			channels.POST("", createChannel(db))
			channels.GET("", listChannels(db))
			channels.GET("/:id", getChannel(db))
			channels.PUT("/:id", updateChannel(db))
			channels.DELETE("/:id", deleteChannel(db))
			channels.GET("/:id/members", listChannelMembers(db))
			channels.POST("/:id/members", addChannelMember(db))
			channels.DELETE("/:id/members/:userId", removeChannelMember(db))
		}
	}

	srv := &http.Server{Addr: ":" + getEnv("PORT", "3003"), Handler: r}
	go func() {
		log.Infof("Channel service starting on port %s", getEnv("PORT", "3003"))
		srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func createChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			WorkspaceID string `json:"workspace_id" binding:"required"`
			Name        string `json:"name" binding:"required"`
			Type        string `json:"type"` // public, private, dm
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		id := generateUUID()
		_, err := db.Exec(`INSERT INTO channels (id, workspace_id, name, type, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW())`,
			id, req.WorkspaceID, req.Name, req.Type, req.Description)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, gin.H{"id": id, "name": req.Name})
	}
}

func listChannels(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Query("workspace_id")
		var channels []map[string]interface{}
		rows, _ := db.Queryx(`SELECT * FROM channels WHERE workspace_id = ? AND deleted_at IS NULL ORDER BY name`, workspaceID)
		for rows.Next() {
			ch := make(map[string]interface{})
			rows.MapScan(ch)
			channels = append(channels, ch)
		}
		c.JSON(200, gin.H{"channels": channels})
	}
}

func getChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ch := make(map[string]interface{})
		row := db.QueryRowx(`SELECT * FROM channels WHERE id = ? AND deleted_at IS NULL`, id)
		if err := row.MapScan(ch); err != nil {
			c.JSON(404, gin.H{"error": "Channel not found"})
			return
		}
		c.JSON(200, ch)
	}
}

func updateChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
			Topic       *string `json:"topic"`
		}
		c.ShouldBindJSON(&req)
		db.Exec(`UPDATE channels SET name = COALESCE(?, name), description = COALESCE(?, description), topic = COALESCE(?, topic), updated_at = NOW() WHERE id = ?`,
			req.Name, req.Description, req.Topic, id)
		c.JSON(200, gin.H{"message": "Updated"})
	}
}

func deleteChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Exec(`UPDATE channels SET deleted_at = NOW() WHERE id = ?`, id)
		c.JSON(204, nil)
	}
}

func listChannelMembers(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Param("id")
		var members []map[string]interface{}
		rows, _ := db.Queryx(`SELECT * FROM channel_members WHERE channel_id = ?`, channelID)
		for rows.Next() {
			m := make(map[string]interface{})
			rows.MapScan(m)
			members = append(members, m)
		}
		c.JSON(200, gin.H{"members": members})
	}
}

func addChannelMember(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Param("id")
		var req struct {
			UserID string `json:"user_id" binding:"required"`
			Role   string `json:"role"`
		}
		c.ShouldBindJSON(&req)
		id := generateUUID()
		db.Exec(`INSERT INTO channel_members (id, channel_id, user_id, role, joined_at) VALUES (?, ?, ?, ?, NOW())`,
			id, channelID, req.UserID, req.Role)
		c.JSON(201, gin.H{"id": id})
	}
}

func removeChannelMember(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Param("id")
		userID := c.Param("userId")
		db.Exec(`DELETE FROM channel_members WHERE channel_id = ? AND user_id = ?`, channelID, userID)
		c.JSON(204, nil)
	}
}

func generateUUID() string {
	return "ch-" + time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
