package models

import (
	"time"
)

// ── Core Channel Models ──

type Channel struct {
	ID          string     `json:"id" db:"id"`
	WorkspaceID string     `json:"workspace_id" db:"workspace_id"`
	Name        string     `json:"name" db:"name"`
	Type        string     `json:"type" db:"type"`
	Description *string    `json:"description" db:"description"`
	Topic       *string    `json:"topic" db:"topic"`
	IconURL     *string    `json:"icon_url" db:"icon_url"`
	IsArchived  bool       `json:"is_archived" db:"is_archived"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type ChannelMember struct {
	ID            string     `json:"id" db:"id"`
	ChannelID     string     `json:"channel_id" db:"channel_id"`
	UserID        string     `json:"user_id" db:"user_id"`
	Role          string     `json:"role" db:"role"`
	Notifications string     `json:"notifications" db:"notifications"`
	JoinedAt      time.Time  `json:"joined_at" db:"joined_at"`
	LastReadAt    *time.Time `json:"last_read_at" db:"last_read_at"`
}

type ChannelPin struct {
	ID        string    `json:"id" db:"id"`
	ChannelID string    `json:"channel_id" db:"channel_id"`
	MessageID string    `json:"message_id" db:"message_id"`
	PinnedBy  string    `json:"pinned_by" db:"pinned_by"`
	PinnedAt  time.Time `json:"pinned_at" db:"pinned_at"`
}

// ── Request DTOs ──

type CreateChannelRequest struct {
	WorkspaceID string  `json:"workspace_id" binding:"required"`
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Type        string  `json:"type" binding:"omitempty,oneof=public private dm group_dm"`
	Description *string `json:"description"`
}

type UpdateChannelRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Topic       *string `json:"topic"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"omitempty,oneof=owner admin member"`
}

// ── Polls ──

type ChannelPoll struct {
	ID          string     `json:"id" db:"id"`
	ChannelID   string     `json:"channel_id" db:"channel_id"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	Question    string     `json:"question" db:"question"`
	IsAnonymous bool       `json:"is_anonymous" db:"is_anonymous"`
	MultiChoice bool       `json:"multi_choice" db:"multi_choice"`
	IsClosed    bool       `json:"is_closed" db:"is_closed"`
	ClosedAt    *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type PollOption struct {
	ID       string `json:"id" db:"id"`
	PollID   string `json:"poll_id" db:"poll_id"`
	Text     string `json:"text" db:"text"`
	Position int    `json:"position" db:"position"`
}

type PollVote struct {
	ID       string    `json:"id" db:"id"`
	PollID   string    `json:"poll_id" db:"poll_id"`
	OptionID string    `json:"option_id" db:"option_id"`
	UserID   string    `json:"user_id" db:"user_id"`
	VotedAt  time.Time `json:"voted_at" db:"voted_at"`
}

type CreatePollRequest struct {
	Question    string   `json:"question" binding:"required,min=1,max=500"`
	Options     []string `json:"options" binding:"required,min=2,max=10"`
	IsAnonymous bool     `json:"is_anonymous"`
	MultiChoice bool     `json:"multi_choice"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type VotePollRequest struct {
	OptionIDs []string `json:"option_ids" binding:"required,min=1"`
}

type PollWithOptions struct {
	ChannelPoll
	Options []PollOption `json:"options"`
}

type PollResult struct {
	OptionID   string `json:"option_id" db:"option_id"`
	OptionText string `json:"option_text" db:"option_text"`
	VoteCount  int    `json:"vote_count" db:"vote_count"`
}

type PollResults struct {
	Poll       ChannelPoll  `json:"poll"`
	Results    []PollResult `json:"results"`
	TotalVotes int          `json:"total_votes"`
}

// ── Scheduled Messages ──

type ScheduledMessage struct {
	ID          string     `json:"id" db:"id"`
	ChannelID   string     `json:"channel_id" db:"channel_id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Content     string     `json:"content" db:"content"`
	ScheduledAt time.Time  `json:"scheduled_at" db:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	Status      string     `json:"status" db:"status"` // pending, sent, cancelled
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateScheduledMessageRequest struct {
	Content     string    `json:"content" binding:"required,min=1"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

type UpdateScheduledMessageRequest struct {
	Content     *string    `json:"content"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

// ── Channel Links / Bridging ──

type ChannelLink struct {
	ID              string    `json:"id" db:"id"`
	SourceChannelID string    `json:"source_channel_id" db:"source_channel_id"`
	TargetChannelID string    `json:"target_channel_id" db:"target_channel_id"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	LinkType        string    `json:"link_type" db:"link_type"` // mirror, forward, sync
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CreateChannelLinkRequest struct {
	TargetChannelID string `json:"target_channel_id" binding:"required"`
	LinkType        string `json:"link_type" binding:"required,oneof=mirror forward sync"`
}

// ── Channel Tabs / Widgets ──

type ChannelTab struct {
	ID        string    `json:"id" db:"id"`
	ChannelID string    `json:"channel_id" db:"channel_id"`
	Name      string    `json:"name" db:"name"`
	TabType   string    `json:"tab_type" db:"tab_type"` // files, links, pinned, custom, notes
	Config    *string   `json:"config" db:"config"`
	Position  int       `json:"position" db:"position"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateTabRequest struct {
	Name    string  `json:"name" binding:"required,min=1,max=50"`
	TabType string  `json:"tab_type" binding:"required,oneof=files links pinned custom notes"`
	Config  *string `json:"config"`
}

type UpdateTabRequest struct {
	Name   *string `json:"name"`
	Config *string `json:"config"`
}

type ReorderTabsRequest struct {
	TabIDs []string `json:"tab_ids" binding:"required,min=1"`
}

// ── Channel Followers ──

type ChannelFollower struct {
	ID        string    `json:"id" db:"id"`
	ChannelID string    `json:"channel_id" db:"channel_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ── Channel Templates ──

type ChannelTemplate struct {
	ID              string    `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     *string   `json:"description" db:"description"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	ChannelType     string    `json:"channel_type" db:"channel_type"`
	DefaultTopic    *string   `json:"default_topic" db:"default_topic"`
	DefaultTabs     *string   `json:"default_tabs" db:"default_tabs"`
	DefaultSettings *string   `json:"default_settings" db:"default_settings"`
	IsPublic        bool      `json:"is_public" db:"is_public"`
	UseCount        int       `json:"use_count" db:"use_count"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CreateTemplateRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description *string `json:"description"`
	IsPublic    bool    `json:"is_public"`
}

type ApplyTemplateRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
	ChannelName string `json:"channel_name" binding:"required,min=1,max=100"`
}
