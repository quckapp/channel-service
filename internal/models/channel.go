package models

import "time"

type Channel struct {
	ID          string     `json:"id" db:"id"`
	WorkspaceID string     `json:"workspace_id" db:"workspace_id"`
	Name        string     `json:"name" db:"name"`
	Type        string     `json:"type" db:"type"`
	Description *string    `json:"description" db:"description"`
	Topic       *string    `json:"topic" db:"topic"`
	IconURL     *string    `json:"icon_url" db:"icon_url"`
	IsArchived  bool       `json:"is_archived" db:"is_archived"`
	CreatedBy   *string    `json:"created_by" db:"created_by"`
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

// Request DTOs

type CreateChannelRequest struct {
	WorkspaceID string  `json:"workspace_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Type        string  `json:"type" binding:"omitempty,oneof=public private dm group_dm"`
	Description *string `json:"description"`
}

type UpdateChannelRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Topic       *string `json:"topic"`
	IconURL     *string `json:"icon_url"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"omitempty,oneof=owner admin member"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=owner admin member"`
}

type UpdateNotificationsRequest struct {
	Notifications string `json:"notifications" binding:"required,oneof=all mentions none"`
}

type UpdateLastReadRequest struct {
	LastReadAt time.Time `json:"last_read_at"`
}

type PinMessageRequest struct {
	MessageID string `json:"message_id" binding:"required"`
}

// Response DTOs

type ChannelResponse struct {
	Channel     *Channel `json:"channel"`
	MemberCount int      `json:"member_count"`
	MyRole      string   `json:"my_role,omitempty"`
}

type ChannelsListResponse struct {
	Channels []*ChannelResponse `json:"channels"`
	Total    int64              `json:"total"`
}

// ── Channel Invites ──

type ChannelInvite struct {
	ID        string     `json:"id" db:"id"`
	ChannelID string     `json:"channel_id" db:"channel_id"`
	CreatedBy string     `json:"created_by" db:"created_by"`
	Code      string     `json:"code" db:"code"`
	MaxUses   int        `json:"max_uses" db:"max_uses"`
	UseCount  int        `json:"use_count" db:"use_count"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type CreateInviteRequest struct {
	MaxUses   int        `json:"max_uses"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type InviteResponse struct {
	Invite      *ChannelInvite `json:"invite"`
	ChannelName string         `json:"channel_name"`
}

// ── Channel Bookmarks ──

type ChannelBookmark struct {
	ID         string    `json:"id" db:"id"`
	ChannelID  string    `json:"channel_id" db:"channel_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	Title      string    `json:"title" db:"title"`
	URL        *string   `json:"url" db:"url"`
	EntityType *string   `json:"entity_type" db:"entity_type"`
	EntityID   *string   `json:"entity_id" db:"entity_id"`
	Position   int       `json:"position" db:"position"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type CreateBookmarkRequest struct {
	Title      string  `json:"title" binding:"required"`
	URL        *string `json:"url"`
	EntityType *string `json:"entity_type"`
	EntityID   *string `json:"entity_id"`
}

type UpdateBookmarkRequest struct {
	Title *string `json:"title"`
	URL   *string `json:"url"`
}

// ── Topic History ──

type TopicHistory struct {
	ID        string    `json:"id" db:"id"`
	ChannelID string    `json:"channel_id" db:"channel_id"`
	OldTopic  *string   `json:"old_topic" db:"old_topic"`
	NewTopic  *string   `json:"new_topic" db:"new_topic"`
	ChangedBy string    `json:"changed_by" db:"changed_by"`
	ChangedAt time.Time `json:"changed_at" db:"changed_at"`
}

// ── Channel Permissions ──

type ChannelPermission struct {
	ID             string    `json:"id" db:"id"`
	ChannelID      string    `json:"channel_id" db:"channel_id"`
	PermissionType string    `json:"permission_type" db:"permission_type"`
	TargetType     string    `json:"target_type" db:"target_type"`
	TargetID       string    `json:"target_id" db:"target_id"`
	Allow          bool      `json:"allow" db:"allow"`
	Deny           bool      `json:"deny" db:"deny"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type SetPermissionRequest struct {
	PermissionType string `json:"permission_type" binding:"required"`
	TargetType     string `json:"target_type" binding:"required,oneof=role user"`
	TargetID       string `json:"target_id" binding:"required"`
	Allow          bool   `json:"allow"`
	Deny           bool   `json:"deny"`
}

type PermissionOverride struct {
	PermissionType string `json:"permission_type"`
	Allow          bool   `json:"allow"`
	Deny           bool   `json:"deny"`
}

// ── Channel Webhooks ──

type ChannelWebhook struct {
	ID              string     `json:"id" db:"id"`
	ChannelID       string     `json:"channel_id" db:"channel_id"`
	Name            string     `json:"name" db:"name"`
	URL             string     `json:"url" db:"url"`
	AvatarURL       *string    `json:"avatar_url" db:"avatar_url"`
	Events          string     `json:"events" db:"events"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	CreatedBy       string     `json:"created_by" db:"created_by"`
	LastTriggeredAt *time.Time `json:"last_triggered_at" db:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateWebhookRequest struct {
	Name      string   `json:"name" binding:"required"`
	URL       string   `json:"url" binding:"required"`
	AvatarURL *string  `json:"avatar_url"`
	Events    []string `json:"events" binding:"required"`
}

type UpdateWebhookRequest struct {
	Name      *string  `json:"name"`
	URL       *string  `json:"url"`
	AvatarURL *string  `json:"avatar_url"`
	Events    []string `json:"events"`
	IsActive  *bool    `json:"is_active"`
}

// ── Channel Reactions ──

type ChannelReaction struct {
	ID        string    `json:"id" db:"id"`
	ChannelID string    `json:"channel_id" db:"channel_id"`
	MessageID string    `json:"message_id" db:"message_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Emoji     string    `json:"emoji" db:"emoji"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AddReactionRequest struct {
	MessageID string `json:"message_id" binding:"required"`
	Emoji     string `json:"emoji" binding:"required"`
}

type RemoveReactionRequest struct {
	MessageID string `json:"message_id" binding:"required"`
	Emoji     string `json:"emoji" binding:"required"`
}

type ReactionSummary struct {
	Emoji string `json:"emoji" db:"emoji"`
	Count int    `json:"count" db:"count"`
}

// ── Channel Moderation ──

type ChannelBan struct {
	ID        string     `json:"id" db:"id"`
	ChannelID string     `json:"channel_id" db:"channel_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	BannedBy  string     `json:"banned_by" db:"banned_by"`
	Reason    *string    `json:"reason" db:"reason"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type ChannelMute struct {
	ID        string     `json:"id" db:"id"`
	ChannelID string     `json:"channel_id" db:"channel_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	MutedBy   string     `json:"muted_by" db:"muted_by"`
	Reason    *string    `json:"reason" db:"reason"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type BanMemberRequest struct {
	Reason    *string    `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type MuteMemberRequest struct {
	Reason    *string    `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type ModerationEntry struct {
	ID        string     `json:"id" db:"id"`
	ChannelID string     `json:"channel_id" db:"channel_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Action    string     `json:"action" db:"action"`
	ActorID   string     `json:"actor_id" db:"actor_id"`
	Reason    *string    `json:"reason" db:"reason"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// ── Channel Announcements ──

type ChannelAnnouncement struct {
	ID        string     `json:"id" db:"id"`
	ChannelID string     `json:"channel_id" db:"channel_id"`
	Title     string     `json:"title" db:"title"`
	Content   string     `json:"content" db:"content"`
	Priority  string     `json:"priority" db:"priority"`
	AuthorID  string     `json:"author_id" db:"author_id"`
	IsPinned  bool       `json:"is_pinned" db:"is_pinned"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateAnnouncementRequest struct {
	Title     string     `json:"title" binding:"required"`
	Content   string     `json:"content" binding:"required"`
	Priority  string     `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type UpdateAnnouncementRequest struct {
	Title     *string    `json:"title"`
	Content   *string    `json:"content"`
	Priority  *string    `json:"priority"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// ── Channel Sections/Categories ──

type ChannelSection struct {
	ID          string    `json:"id" db:"id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Position    int       `json:"position" db:"position"`
	IsCollapsed bool      `json:"is_collapsed" db:"is_collapsed"`
	ChannelIDs  string    `json:"channel_ids" db:"channel_ids"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateSectionRequest struct {
	WorkspaceID string   `json:"workspace_id" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	ChannelIDs  []string `json:"channel_ids"`
}

type UpdateSectionRequest struct {
	Name        *string  `json:"name"`
	IsCollapsed *bool    `json:"is_collapsed"`
	ChannelIDs  []string `json:"channel_ids"`
}

// ── Channel Analytics ──

type ChannelStats struct {
	MemberCount       int `json:"member_count"`
	PinCount          int `json:"pin_count"`
	MessageCountToday int `json:"message_count_today"`
	ActiveMembersWeek int `json:"active_members_week"`
}

type ChannelActivity struct {
	Date         string `json:"date" db:"date"`
	MessageCount int    `json:"message_count" db:"message_count"`
	ActiveUsers  int    `json:"active_users" db:"active_users"`
}

type MostActiveMember struct {
	UserID       string `json:"user_id" db:"user_id"`
	MessageCount int    `json:"message_count" db:"message_count"`
}

// ── Misc DTOs ──

type TransferOwnershipRequest struct {
	NewOwnerID string `json:"new_owner_id" binding:"required"`
}

type BulkAddMembersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
	Role    string   `json:"role" binding:"omitempty,oneof=admin member"`
}

type CloneChannelRequest struct {
	Name            string `json:"name" binding:"required"`
	IncludeMembers  bool   `json:"include_members"`
	IncludePins     bool   `json:"include_pins"`
	IncludeSettings bool   `json:"include_settings"`
}

// ── Channel Threads ──

type ChannelThread struct {
	ID          string     `json:"id" db:"id"`
	ChannelID   string     `json:"channel_id" db:"channel_id"`
	MessageID   string     `json:"message_id" db:"message_id"`
	Title       *string    `json:"title" db:"title"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	IsLocked    bool       `json:"is_locked" db:"is_locked"`
	IsResolved  bool       `json:"is_resolved" db:"is_resolved"`
	ReplyCount  int        `json:"reply_count" db:"reply_count"`
	LastReplyAt *time.Time `json:"last_reply_at" db:"last_reply_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type ThreadReply struct {
	ID        string    `json:"id" db:"id"`
	ThreadID  string    `json:"thread_id" db:"thread_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	ParentID  *string   `json:"parent_id" db:"parent_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateThreadRequest struct {
	MessageID string  `json:"message_id" binding:"required"`
	Title     *string `json:"title"`
}

type CreateReplyRequest struct {
	Content  string  `json:"content" binding:"required"`
	ParentID *string `json:"parent_id"`
}

type UpdateReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

type ThreadFollower struct {
	ID        string    `json:"id" db:"id"`
	ThreadID  string    `json:"thread_id" db:"thread_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ── Channel Settings ──

type ChannelSetting struct {
	ID                  string    `json:"id" db:"id"`
	ChannelID           string    `json:"channel_id" db:"channel_id"`
	SlowModeInterval    int       `json:"slow_mode_interval" db:"slow_mode_interval"`
	MaxPins             int       `json:"max_pins" db:"max_pins"`
	MaxBookmarks        int       `json:"max_bookmarks" db:"max_bookmarks"`
	AllowThreads        bool      `json:"allow_threads" db:"allow_threads"`
	AllowReactions      bool      `json:"allow_reactions" db:"allow_reactions"`
	AllowInvites        bool      `json:"allow_invites" db:"allow_invites"`
	AutoArchiveDays     int       `json:"auto_archive_days" db:"auto_archive_days"`
	DefaultNotification string    `json:"default_notification" db:"default_notification"`
	CustomEmoji         bool      `json:"custom_emoji" db:"custom_emoji"`
	LinkPreviews        bool      `json:"link_previews" db:"link_previews"`
	MemberLimit         int       `json:"member_limit" db:"member_limit"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateSettingsRequest struct {
	SlowModeInterval    *int    `json:"slow_mode_interval"`
	MaxPins             *int    `json:"max_pins"`
	MaxBookmarks        *int    `json:"max_bookmarks"`
	AllowThreads        *bool   `json:"allow_threads"`
	AllowReactions      *bool   `json:"allow_reactions"`
	AllowInvites        *bool   `json:"allow_invites"`
	AutoArchiveDays     *int    `json:"auto_archive_days"`
	DefaultNotification *string `json:"default_notification"`
	CustomEmoji         *bool   `json:"custom_emoji"`
	LinkPreviews        *bool   `json:"link_previews"`
	MemberLimit         *int    `json:"member_limit"`
}

// ── Starred Channels ──

type StarredChannel struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	ChannelID string    `json:"channel_id" db:"channel_id"`
	Position  int       `json:"position" db:"position"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ── Read Receipts ──

type ReadReceipt struct {
	ID        string    `json:"id" db:"id"`
	ChannelID string    `json:"channel_id" db:"channel_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	MessageID string    `json:"message_id" db:"message_id"`
	ReadAt    time.Time `json:"read_at" db:"read_at"`
}

type MarkReadRequest struct {
	MessageID string `json:"message_id" binding:"required"`
}

type ReadReceiptSummary struct {
	MessageID string   `json:"message_id"`
	ReadBy    []string `json:"read_by"`
	ReadCount int      `json:"read_count"`
}

// ── Scheduled Messages ──

type ScheduledMessage struct {
	ID          string     `json:"id" db:"id"`
	ChannelID   string     `json:"channel_id" db:"channel_id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Content     string     `json:"content" db:"content"`
	ScheduledAt time.Time  `json:"scheduled_at" db:"scheduled_at"`
	Status      string     `json:"status" db:"status"`
	SentAt      *time.Time `json:"sent_at" db:"sent_at"`
	ThreadID    *string    `json:"thread_id" db:"thread_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateScheduledMessageRequest struct {
	Content     string  `json:"content" binding:"required"`
	ScheduledAt string  `json:"scheduled_at" binding:"required"`
	ThreadID    *string `json:"thread_id"`
}

type UpdateScheduledMessageRequest struct {
	Content     *string `json:"content"`
	ScheduledAt *string `json:"scheduled_at"`
}

// ── Channel Activity Log ──

type ChannelActivityLog struct {
	ID        string    `json:"id" db:"id"`
	ChannelID string    `json:"channel_id" db:"channel_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	TargetID  *string   `json:"target_id" db:"target_id"`
	Details   *string   `json:"details" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ── Channel Templates ──

type ChannelTemplate struct {
	ID          string    `json:"id" db:"id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"`
	Topic       *string   `json:"topic" db:"topic"`
	Settings    *string   `json:"settings" db:"settings"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateTemplateRequest struct {
	WorkspaceID string  `json:"workspace_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Type        string  `json:"type" binding:"omitempty,oneof=public private"`
	Topic       *string `json:"topic"`
	Settings    *string `json:"settings"`
}

type UpdateTemplateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	Topic       *string `json:"topic"`
	Settings    *string `json:"settings"`
}

type ApplyTemplateRequest struct {
	ChannelName string `json:"channel_name" binding:"required"`
}

// ── Voice Channels ──

type VoiceChannelState struct {
	ID             string     `json:"id" db:"id"`
	ChannelID      string     `json:"channel_id" db:"channel_id"`
	UserID         string     `json:"user_id" db:"user_id"`
	IsMuted        bool       `json:"is_muted" db:"is_muted"`
	IsDeafened     bool       `json:"is_deafened" db:"is_deafened"`
	IsScreenShare  bool       `json:"is_screen_share" db:"is_screen_share"`
	IsVideoOn      bool       `json:"is_video_on" db:"is_video_on"`
	JoinedAt       time.Time  `json:"joined_at" db:"joined_at"`
	DisconnectedAt *time.Time `json:"disconnected_at" db:"disconnected_at"`
}

type UpdateVoiceStateRequest struct {
	IsMuted       *bool `json:"is_muted"`
	IsDeafened    *bool `json:"is_deafened"`
	IsScreenShare *bool `json:"is_screen_share"`
	IsVideoOn     *bool `json:"is_video_on"`
}

type VoiceParticipant struct {
	UserID        string `json:"user_id"`
	IsMuted       bool   `json:"is_muted"`
	IsDeafened    bool   `json:"is_deafened"`
	IsScreenShare bool   `json:"is_screen_share"`
	IsVideoOn     bool   `json:"is_video_on"`
}

// ── Channel Followers ──

type ChannelFollower struct {
	ID         string    `json:"id" db:"id"`
	ChannelID  string    `json:"channel_id" db:"channel_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	FollowedAt time.Time `json:"followed_at" db:"followed_at"`
}

// ── Bulk Operations Extended ──

type BulkDeleteMembersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
}

type BulkUpdateRoleRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
	Role    string   `json:"role" binding:"required,oneof=admin member"`
}

type BulkActionResult struct {
	Successful int      `json:"successful"`
	Failed     int      `json:"failed"`
	Errors     []string `json:"errors,omitempty"`
}
