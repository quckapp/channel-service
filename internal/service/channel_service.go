package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/quckapp/channel-service/internal/db"
	"github.com/quckapp/channel-service/internal/models"
	"github.com/quckapp/channel-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	ErrChannelNotFound      = errors.New("channel not found")
	ErrNotMember            = errors.New("not a member of this channel")
	ErrNotAuthorized        = errors.New("not authorized")
	ErrAlreadyMember        = errors.New("already a member")
	ErrAlreadyPinned        = errors.New("message already pinned")
	ErrPinNotFound          = errors.New("pin not found")
	ErrChannelArchived      = errors.New("channel is archived")
	ErrCannotLeaveOwner     = errors.New("owner cannot leave channel, transfer ownership first")
	ErrInviteNotFound       = errors.New("invite not found")
	ErrInviteExpired        = errors.New("invite has expired")
	ErrInviteMaxUses        = errors.New("invite has reached max uses")
	ErrBookmarkNotFound     = errors.New("bookmark not found")
	ErrBookmarkLimitReached = errors.New("bookmark limit reached")
	ErrPermissionNotFound   = errors.New("permission not found")
	ErrWebhookNotFound      = errors.New("webhook not found")
	ErrReactionExists       = errors.New("reaction already exists")
	ErrReactionNotFound     = errors.New("reaction not found")
	ErrUserBanned           = errors.New("user is banned from this channel")
	ErrUserMuted            = errors.New("user is muted in this channel")
	ErrAnnouncementNotFound      = errors.New("announcement not found")
	ErrSectionNotFound           = errors.New("section not found")
	ErrThreadNotFound            = errors.New("thread not found")
	ErrReplyNotFound             = errors.New("reply not found")
	ErrSettingsNotFound          = errors.New("channel settings not found")
	ErrScheduledMessageNotFound  = errors.New("scheduled message not found")
	ErrTemplateNotFound          = errors.New("template not found")
	ErrAlreadyStarred            = errors.New("channel already starred")
	ErrNotStarred                = errors.New("channel not starred")
	ErrNotInVoice                = errors.New("not in voice channel")
	ErrAlreadyInVoice            = errors.New("already in voice channel")
	ErrAlreadyFollowing          = errors.New("already following channel")
	ErrNotFollowing              = errors.New("not following channel")
)

const (
	cacheTTL          = 10 * time.Minute
	cacheKeyChannel   = "channel:%s"
	cacheKeyMembers   = "channel:%s:members"
)

type ChannelService struct {
	channelRepo          *repository.ChannelRepository
	memberRepo           *repository.MemberRepository
	pinRepo              *repository.PinRepository
	inviteRepo           *repository.InviteRepository
	bookmarkRepo         *repository.BookmarkRepository
	topicHistoryRepo     *repository.TopicHistoryRepository
	permissionRepo       *repository.PermissionRepository
	webhookRepo          *repository.WebhookRepository
	reactionRepo         *repository.ReactionRepository
	moderationRepo       *repository.ModerationRepository
	announcementRepo     *repository.AnnouncementRepository
	sectionRepo          *repository.SectionRepository
	analyticsRepo        *repository.AnalyticsRepository
	threadRepo           *repository.ThreadRepository
	settingsRepo         *repository.SettingsRepository
	starredRepo          *repository.StarredRepository
	readReceiptRepo      *repository.ReadReceiptRepository
	scheduledMessageRepo *repository.ScheduledMessageRepository
	activityLogRepo      *repository.ActivityLogRepository
	templateRepo         *repository.TemplateRepository
	voiceRepo            *repository.VoiceRepository
	followerRepo         *repository.FollowerRepository
	redis                *redis.Client
	kafka                *db.KafkaProducer
	logger               *logrus.Logger
}

func NewChannelService(
	channelRepo *repository.ChannelRepository,
	memberRepo *repository.MemberRepository,
	pinRepo *repository.PinRepository,
	inviteRepo *repository.InviteRepository,
	bookmarkRepo *repository.BookmarkRepository,
	topicHistoryRepo *repository.TopicHistoryRepository,
	permissionRepo *repository.PermissionRepository,
	webhookRepo *repository.WebhookRepository,
	reactionRepo *repository.ReactionRepository,
	moderationRepo *repository.ModerationRepository,
	announcementRepo *repository.AnnouncementRepository,
	sectionRepo *repository.SectionRepository,
	analyticsRepo *repository.AnalyticsRepository,
	threadRepo *repository.ThreadRepository,
	settingsRepo *repository.SettingsRepository,
	starredRepo *repository.StarredRepository,
	readReceiptRepo *repository.ReadReceiptRepository,
	scheduledMessageRepo *repository.ScheduledMessageRepository,
	activityLogRepo *repository.ActivityLogRepository,
	templateRepo *repository.TemplateRepository,
	voiceRepo *repository.VoiceRepository,
	followerRepo *repository.FollowerRepository,
	redis *redis.Client,
	kafka *db.KafkaProducer,
	logger *logrus.Logger,
) *ChannelService {
	return &ChannelService{
		channelRepo:          channelRepo,
		memberRepo:           memberRepo,
		pinRepo:              pinRepo,
		inviteRepo:           inviteRepo,
		bookmarkRepo:         bookmarkRepo,
		topicHistoryRepo:     topicHistoryRepo,
		permissionRepo:       permissionRepo,
		webhookRepo:          webhookRepo,
		reactionRepo:         reactionRepo,
		moderationRepo:       moderationRepo,
		announcementRepo:     announcementRepo,
		sectionRepo:          sectionRepo,
		analyticsRepo:        analyticsRepo,
		threadRepo:           threadRepo,
		settingsRepo:         settingsRepo,
		starredRepo:          starredRepo,
		readReceiptRepo:      readReceiptRepo,
		scheduledMessageRepo: scheduledMessageRepo,
		activityLogRepo:      activityLogRepo,
		templateRepo:         templateRepo,
		voiceRepo:            voiceRepo,
		followerRepo:         followerRepo,
		redis:                redis,
		kafka:                kafka,
		logger:               logger,
	}
}

// ── Channel CRUD ──

func (s *ChannelService) CreateChannel(ctx context.Context, userID string, req *models.CreateChannelRequest) (*models.Channel, error) {
	chType := req.Type
	if chType == "" {
		chType = "public"
	}

	ch := &models.Channel{
		ID:          uuid.New().String(),
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Type:        chType,
		Description: req.Description,
		CreatedBy:   &userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.channelRepo.Create(ctx, ch); err != nil {
		return nil, err
	}

	// Add creator as owner
	member := &models.ChannelMember{
		ID:            uuid.New().String(),
		ChannelID:     ch.ID,
		UserID:        userID,
		Role:          "owner",
		Notifications: "all",
		JoinedAt:      time.Now(),
	}
	s.memberRepo.Create(ctx, member)

	s.publishEvent(ctx, "channel-events", ch.ID, "channel.created", map[string]interface{}{
		"channel":      ch,
		"workspace_id": ch.WorkspaceID,
	})

	return ch, nil
}

func (s *ChannelService) GetChannel(ctx context.Context, id, userID string) (*models.ChannelResponse, error) {
	// Try cache
	if cached, err := s.getCachedChannel(ctx, id); err == nil && cached != nil {
		memberCount, _ := s.channelRepo.GetMemberCount(ctx, id)
		role, _ := s.memberRepo.GetRole(ctx, id, userID)
		return &models.ChannelResponse{Channel: cached, MemberCount: memberCount, MyRole: role}, nil
	}

	ch, err := s.channelRepo.GetByID(ctx, id)
	if err != nil || ch == nil {
		return nil, ErrChannelNotFound
	}

	memberCount, _ := s.channelRepo.GetMemberCount(ctx, id)
	role, _ := s.memberRepo.GetRole(ctx, id, userID)

	s.cacheChannel(ctx, id, ch)
	return &models.ChannelResponse{Channel: ch, MemberCount: memberCount, MyRole: role}, nil
}

func (s *ChannelService) UpdateChannel(ctx context.Context, id, userID string, req *models.UpdateChannelRequest) (*models.Channel, error) {
	ch, err := s.channelRepo.GetByID(ctx, id)
	if err != nil || ch == nil {
		return nil, ErrChannelNotFound
	}

	role, _ := s.memberRepo.GetRole(ctx, id, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		ch.Name = *req.Name
	}
	if req.Description != nil {
		ch.Description = req.Description
	}
	if req.Topic != nil {
		// Record topic history
		oldTopic := ch.Topic
		s.topicHistoryRepo.Create(ctx, &models.TopicHistory{
			ID:        uuid.New().String(),
			ChannelID: id,
			OldTopic:  oldTopic,
			NewTopic:  req.Topic,
			ChangedBy: userID,
			ChangedAt: time.Now(),
		})
		ch.Topic = req.Topic
	}
	if req.IconURL != nil {
		ch.IconURL = req.IconURL
	}

	if err := s.channelRepo.Update(ctx, ch); err != nil {
		return nil, err
	}

	s.invalidateChannel(ctx, id)
	s.publishEvent(ctx, "channel-events", id, "channel.updated", map[string]interface{}{
		"channel":    ch,
		"updated_by": userID,
	})

	return ch, nil
}

func (s *ChannelService) DeleteChannel(ctx context.Context, id, userID string) error {
	ch, err := s.channelRepo.GetByID(ctx, id)
	if err != nil || ch == nil {
		return ErrChannelNotFound
	}

	role, _ := s.memberRepo.GetRole(ctx, id, userID)
	if role != "owner" {
		return ErrNotAuthorized
	}

	if err := s.channelRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.invalidateChannel(ctx, id)
	s.publishEvent(ctx, "channel-events", id, "channel.deleted", map[string]interface{}{
		"channel_id":   id,
		"workspace_id": ch.WorkspaceID,
		"deleted_by":   userID,
	})

	return nil
}

func (s *ChannelService) ListChannels(ctx context.Context, workspaceID string) ([]*models.Channel, error) {
	return s.channelRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *ChannelService) ListUserChannels(ctx context.Context, userID, workspaceID string) ([]*models.Channel, error) {
	return s.channelRepo.ListByUser(ctx, userID, workspaceID)
}

func (s *ChannelService) SearchChannels(ctx context.Context, workspaceID, query string) ([]*models.Channel, error) {
	return s.channelRepo.Search(ctx, workspaceID, query)
}

// ── Archive ──

func (s *ChannelService) ArchiveChannel(ctx context.Context, id, userID string) error {
	ch, err := s.channelRepo.GetByID(ctx, id)
	if err != nil || ch == nil {
		return ErrChannelNotFound
	}

	role, _ := s.memberRepo.GetRole(ctx, id, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	if err := s.channelRepo.Archive(ctx, id); err != nil {
		return err
	}

	s.invalidateChannel(ctx, id)
	s.publishEvent(ctx, "channel-events", id, "channel.archived", map[string]interface{}{
		"channel_id":   id,
		"archived_by":  userID,
	})

	return nil
}

func (s *ChannelService) UnarchiveChannel(ctx context.Context, id, userID string) error {
	ch, err := s.channelRepo.GetByID(ctx, id)
	if err != nil || ch == nil {
		return ErrChannelNotFound
	}

	role, _ := s.memberRepo.GetRole(ctx, id, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	if err := s.channelRepo.Unarchive(ctx, id); err != nil {
		return err
	}

	s.invalidateChannel(ctx, id)
	s.publishEvent(ctx, "channel-events", id, "channel.unarchived", map[string]interface{}{
		"channel_id":     id,
		"unarchived_by":  userID,
	})

	return nil
}

// ── Member Management ──

func (s *ChannelService) AddMember(ctx context.Context, channelID, userID string, req *models.AddMemberRequest) (*models.ChannelMember, error) {
	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil || ch == nil {
		return nil, ErrChannelNotFound
	}
	if ch.IsArchived {
		return nil, ErrChannelArchived
	}

	// Check if user is banned
	banned, _ := s.moderationRepo.IsBanned(ctx, channelID, req.UserID)
	if banned {
		return nil, ErrUserBanned
	}

	existing, _ := s.memberRepo.GetByChannelAndUser(ctx, channelID, req.UserID)
	if existing != nil {
		return nil, ErrAlreadyMember
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	member := &models.ChannelMember{
		ID:            uuid.New().String(),
		ChannelID:     channelID,
		UserID:        req.UserID,
		Role:          role,
		Notifications: "all",
		JoinedAt:      time.Now(),
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	s.invalidateChannel(ctx, channelID)
	s.publishEvent(ctx, "channel-events", channelID, "member.joined", map[string]interface{}{
		"channel_id": channelID,
		"user_id":    req.UserID,
		"added_by":   userID,
	})

	return member, nil
}

func (s *ChannelService) RemoveMember(ctx context.Context, channelID, memberUserID, requestorID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, requestorID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	memberRole, _ := s.memberRepo.GetRole(ctx, channelID, memberUserID)
	if memberRole == "owner" {
		return ErrNotAuthorized
	}

	if err := s.memberRepo.Remove(ctx, channelID, memberUserID); err != nil {
		return err
	}

	s.invalidateChannel(ctx, channelID)
	s.publishEvent(ctx, "channel-events", channelID, "member.removed", map[string]interface{}{
		"channel_id": channelID,
		"user_id":    memberUserID,
		"removed_by": requestorID,
	})

	return nil
}

func (s *ChannelService) LeaveChannel(ctx context.Context, channelID, userID string) error {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return ErrNotMember
	}

	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role == "owner" {
		return ErrCannotLeaveOwner
	}

	if err := s.memberRepo.Remove(ctx, channelID, userID); err != nil {
		return err
	}

	s.invalidateChannel(ctx, channelID)
	s.publishEvent(ctx, "channel-events", channelID, "member.left", map[string]interface{}{
		"channel_id": channelID,
		"user_id":    userID,
	})

	return nil
}

func (s *ChannelService) GetMember(ctx context.Context, channelID, userID string) (*models.ChannelMember, error) {
	member, err := s.memberRepo.GetByChannelAndUser(ctx, channelID, userID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}
	return member, nil
}

func (s *ChannelService) ListMembers(ctx context.Context, channelID string) ([]*models.ChannelMember, error) {
	return s.memberRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) UpdateMemberRole(ctx context.Context, channelID, memberUserID, requestorID, newRole string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, requestorID)
	if role != "owner" {
		return ErrNotAuthorized
	}

	if err := s.memberRepo.UpdateRole(ctx, channelID, memberUserID, newRole); err != nil {
		return err
	}

	s.publishEvent(ctx, "channel-events", channelID, "member.role_updated", map[string]interface{}{
		"channel_id": channelID,
		"user_id":    memberUserID,
		"new_role":   newRole,
	})

	return nil
}

func (s *ChannelService) UpdateNotifications(ctx context.Context, channelID, userID, level string) error {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return ErrNotMember
	}

	return s.memberRepo.UpdateNotifications(ctx, channelID, userID, level)
}

func (s *ChannelService) UpdateLastRead(ctx context.Context, channelID, userID string) error {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return ErrNotMember
	}

	return s.memberRepo.UpdateLastRead(ctx, channelID, userID)
}

// ── Pins ──

func (s *ChannelService) PinMessage(ctx context.Context, channelID, userID string, req *models.PinMessageRequest) (*models.ChannelPin, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role == "" {
		return nil, ErrNotMember
	}

	existing, _ := s.pinRepo.GetByChannelAndMessage(ctx, channelID, req.MessageID)
	if existing != nil {
		return nil, ErrAlreadyPinned
	}

	pin := &models.ChannelPin{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		MessageID: req.MessageID,
		PinnedBy:  userID,
		PinnedAt:  time.Now(),
	}

	if err := s.pinRepo.Create(ctx, pin); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "channel-events", channelID, "message.pinned", map[string]interface{}{
		"channel_id": channelID,
		"message_id": req.MessageID,
		"pinned_by":  userID,
	})

	return pin, nil
}

func (s *ChannelService) UnpinMessage(ctx context.Context, channelID, messageID, userID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role == "" {
		return ErrNotMember
	}

	existing, _ := s.pinRepo.GetByChannelAndMessage(ctx, channelID, messageID)
	if existing == nil {
		return ErrPinNotFound
	}

	if err := s.pinRepo.Delete(ctx, channelID, messageID); err != nil {
		return err
	}

	s.publishEvent(ctx, "channel-events", channelID, "message.unpinned", map[string]interface{}{
		"channel_id": channelID,
		"message_id": messageID,
		"unpinned_by": userID,
	})

	return nil
}

func (s *ChannelService) ListPins(ctx context.Context, channelID string) ([]*models.ChannelPin, error) {
	return s.pinRepo.ListByChannel(ctx, channelID)
}

// ── Typing ──

func (s *ChannelService) SetTyping(ctx context.Context, channelID, userID string) error {
	if s.redis == nil {
		return nil
	}
	key := fmt.Sprintf("typing:%s:%s", channelID, userID)
	s.redis.Set(ctx, key, "1", 5*time.Second)

	s.publishEvent(ctx, "channel-events", channelID, "typing.started", map[string]interface{}{
		"channel_id": channelID,
		"user_id":    userID,
	})

	return nil
}

func (s *ChannelService) GetTyping(ctx context.Context, channelID string) ([]string, error) {
	if s.redis == nil {
		return nil, nil
	}
	pattern := fmt.Sprintf("typing:%s:*", channelID)
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	var users []string
	for _, key := range keys {
		// key format: typing:<channelID>:<userID>
		if len(key) > len(pattern)-1 {
			userID := key[len(fmt.Sprintf("typing:%s:", channelID)):]
			users = append(users, userID)
		}
	}
	return users, nil
}

// ── Invites ──

func (s *ChannelService) CreateInvite(ctx context.Context, channelID, userID string, req *models.CreateInviteRequest) (*models.ChannelInvite, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	invite := &models.ChannelInvite{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		CreatedBy: userID,
		Code:      uuid.New().String()[:8],
		MaxUses:   req.MaxUses,
		UseCount:  0,
		ExpiresAt: req.ExpiresAt,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *ChannelService) ListInvites(ctx context.Context, channelID, userID string) ([]*models.ChannelInvite, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}
	return s.inviteRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) DeleteInvite(ctx context.Context, channelID, inviteID, userID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	invite, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil || invite == nil {
		return ErrInviteNotFound
	}

	return s.inviteRepo.Delete(ctx, inviteID)
}

func (s *ChannelService) JoinByCode(ctx context.Context, code, userID string) (*models.ChannelMember, error) {
	invite, err := s.inviteRepo.GetByCode(ctx, code)
	if err != nil || invite == nil {
		return nil, ErrInviteNotFound
	}

	if !invite.IsActive {
		return nil, ErrInviteExpired
	}
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now()) {
		return nil, ErrInviteExpired
	}
	if invite.MaxUses > 0 && invite.UseCount >= invite.MaxUses {
		return nil, ErrInviteMaxUses
	}

	// Check if banned
	banned, _ := s.moderationRepo.IsBanned(ctx, invite.ChannelID, userID)
	if banned {
		return nil, ErrUserBanned
	}

	existing, _ := s.memberRepo.GetByChannelAndUser(ctx, invite.ChannelID, userID)
	if existing != nil {
		return nil, ErrAlreadyMember
	}

	member := &models.ChannelMember{
		ID:            uuid.New().String(),
		ChannelID:     invite.ChannelID,
		UserID:        userID,
		Role:          "member",
		Notifications: "all",
		JoinedAt:      time.Now(),
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	s.inviteRepo.IncrementUseCount(ctx, invite.ID)
	s.invalidateChannel(ctx, invite.ChannelID)

	return member, nil
}

// ── Bookmarks ──

func (s *ChannelService) CreateBookmark(ctx context.Context, channelID, userID string, req *models.CreateBookmarkRequest) (*models.ChannelBookmark, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	count, _ := s.bookmarkRepo.CountByUser(ctx, channelID, userID)
	if count >= 100 {
		return nil, ErrBookmarkLimitReached
	}

	maxPos, _ := s.bookmarkRepo.GetMaxPosition(ctx, channelID, userID)
	bookmark := &models.ChannelBookmark{
		ID:         uuid.New().String(),
		ChannelID:  channelID,
		UserID:     userID,
		Title:      req.Title,
		URL:        req.URL,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Position:   maxPos + 1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.bookmarkRepo.Create(ctx, bookmark); err != nil {
		return nil, err
	}

	return bookmark, nil
}

func (s *ChannelService) ListBookmarks(ctx context.Context, channelID, userID string) ([]*models.ChannelBookmark, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.bookmarkRepo.ListByUserInChannel(ctx, channelID, userID)
}

func (s *ChannelService) UpdateBookmark(ctx context.Context, channelID, bookmarkID, userID string, req *models.UpdateBookmarkRequest) (*models.ChannelBookmark, error) {
	bookmark, err := s.bookmarkRepo.GetByID(ctx, bookmarkID)
	if err != nil || bookmark == nil {
		return nil, ErrBookmarkNotFound
	}
	if bookmark.UserID != userID {
		return nil, ErrNotAuthorized
	}

	if req.Title != nil {
		bookmark.Title = *req.Title
	}
	if req.URL != nil {
		bookmark.URL = req.URL
	}

	if err := s.bookmarkRepo.Update(ctx, bookmark); err != nil {
		return nil, err
	}

	return bookmark, nil
}

func (s *ChannelService) DeleteBookmark(ctx context.Context, channelID, bookmarkID, userID string) error {
	bookmark, err := s.bookmarkRepo.GetByID(ctx, bookmarkID)
	if err != nil || bookmark == nil {
		return ErrBookmarkNotFound
	}
	if bookmark.UserID != userID {
		return ErrNotAuthorized
	}

	return s.bookmarkRepo.Delete(ctx, bookmarkID)
}

// ── Topic History ──

func (s *ChannelService) GetTopicHistory(ctx context.Context, channelID, userID string, limit, offset int) ([]*models.TopicHistory, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	if limit <= 0 {
		limit = 20
	}
	return s.topicHistoryRepo.ListByChannel(ctx, channelID, limit, offset)
}

// ── Permissions ──

func (s *ChannelService) SetPermission(ctx context.Context, channelID, userID string, req *models.SetPermissionRequest) (*models.ChannelPermission, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	perm := &models.ChannelPermission{
		ID:             uuid.New().String(),
		ChannelID:      channelID,
		PermissionType: req.PermissionType,
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		Allow:          req.Allow,
		Deny:           req.Deny,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.permissionRepo.Set(ctx, perm); err != nil {
		return nil, err
	}

	return perm, nil
}

func (s *ChannelService) ListPermissions(ctx context.Context, channelID, userID string) ([]*models.ChannelPermission, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.permissionRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) DeletePermission(ctx context.Context, channelID, permissionID, userID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	perm, err := s.permissionRepo.GetByID(ctx, permissionID)
	if err != nil || perm == nil {
		return ErrPermissionNotFound
	}

	return s.permissionRepo.Delete(ctx, permissionID)
}

// ── Webhooks ──

func (s *ChannelService) CreateWebhook(ctx context.Context, channelID, userID string, req *models.CreateWebhookRequest) (*models.ChannelWebhook, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	eventsJSON, _ := json.Marshal(req.Events)
	webhook := &models.ChannelWebhook{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		Name:      req.Name,
		URL:       req.URL,
		AvatarURL: req.AvatarURL,
		Events:    string(eventsJSON),
		IsActive:  true,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.webhookRepo.Create(ctx, webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

func (s *ChannelService) ListWebhooks(ctx context.Context, channelID, userID string) ([]*models.ChannelWebhook, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}
	return s.webhookRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) UpdateWebhook(ctx context.Context, channelID, webhookID, userID string, req *models.UpdateWebhookRequest) (*models.ChannelWebhook, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil || webhook == nil {
		return nil, ErrWebhookNotFound
	}

	if req.Name != nil {
		webhook.Name = *req.Name
	}
	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.AvatarURL != nil {
		webhook.AvatarURL = req.AvatarURL
	}
	if req.Events != nil {
		eventsJSON, _ := json.Marshal(req.Events)
		webhook.Events = string(eventsJSON)
	}
	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if err := s.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

func (s *ChannelService) DeleteWebhook(ctx context.Context, channelID, webhookID, userID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil || webhook == nil {
		return ErrWebhookNotFound
	}

	return s.webhookRepo.Delete(ctx, webhookID)
}

func (s *ChannelService) TestWebhook(ctx context.Context, channelID, webhookID, userID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil || webhook == nil {
		return ErrWebhookNotFound
	}

	s.webhookRepo.UpdateLastTriggered(ctx, webhookID)
	s.publishEvent(ctx, "channel-events", channelID, "webhook.tested", map[string]interface{}{
		"webhook_id": webhookID,
		"channel_id": channelID,
		"tested_by":  userID,
	})

	return nil
}

// ── Reactions ──

func (s *ChannelService) AddReaction(ctx context.Context, channelID, userID string, req *models.AddReactionRequest) (*models.ChannelReaction, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	exists, _ := s.reactionRepo.Exists(ctx, channelID, req.MessageID, userID, req.Emoji)
	if exists {
		return nil, ErrReactionExists
	}

	reaction := &models.ChannelReaction{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		MessageID: req.MessageID,
		UserID:    userID,
		Emoji:     req.Emoji,
		CreatedAt: time.Now(),
	}

	if err := s.reactionRepo.Create(ctx, reaction); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "channel-events", channelID, "reaction.added", map[string]interface{}{
		"channel_id": channelID,
		"message_id": req.MessageID,
		"emoji":      req.Emoji,
		"user_id":    userID,
	})

	return reaction, nil
}

func (s *ChannelService) RemoveReaction(ctx context.Context, channelID, userID string, req *models.RemoveReactionRequest) error {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return ErrNotMember
	}

	exists, _ := s.reactionRepo.Exists(ctx, channelID, req.MessageID, userID, req.Emoji)
	if !exists {
		return ErrReactionNotFound
	}

	return s.reactionRepo.Delete(ctx, channelID, req.MessageID, userID, req.Emoji)
}

func (s *ChannelService) ListReactions(ctx context.Context, channelID, messageID string) ([]*models.ChannelReaction, error) {
	return s.reactionRepo.ListByMessage(ctx, channelID, messageID)
}

func (s *ChannelService) GetReactionSummary(ctx context.Context, channelID, messageID string) ([]models.ReactionSummary, error) {
	return s.reactionRepo.GetSummary(ctx, channelID, messageID)
}

// ── Moderation ──

func (s *ChannelService) BanMember(ctx context.Context, channelID, targetUserID, userID string, req *models.BanMemberRequest) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	targetRole, _ := s.memberRepo.GetRole(ctx, channelID, targetUserID)
	if targetRole == "owner" {
		return ErrNotAuthorized
	}

	ban := &models.ChannelBan{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    targetUserID,
		BannedBy:  userID,
		Reason:    req.Reason,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.moderationRepo.CreateBan(ctx, ban); err != nil {
		return err
	}

	// Remove the member
	s.memberRepo.Remove(ctx, channelID, targetUserID)

	// Log moderation action
	s.moderationRepo.LogAction(ctx, &models.ModerationEntry{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    targetUserID,
		Action:    "ban",
		ActorID:   userID,
		Reason:    req.Reason,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
	})

	s.invalidateChannel(ctx, channelID)
	s.publishEvent(ctx, "channel-events", channelID, "member.banned", map[string]interface{}{
		"channel_id": channelID,
		"user_id":    targetUserID,
		"banned_by":  userID,
	})

	return nil
}

func (s *ChannelService) UnbanMember(ctx context.Context, channelID, targetUserID, userID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	if err := s.moderationRepo.RemoveBan(ctx, channelID, targetUserID); err != nil {
		return err
	}

	s.moderationRepo.LogAction(ctx, &models.ModerationEntry{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    targetUserID,
		Action:    "unban",
		ActorID:   userID,
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *ChannelService) MuteMember(ctx context.Context, channelID, targetUserID, userID string, req *models.MuteMemberRequest) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	targetRole, _ := s.memberRepo.GetRole(ctx, channelID, targetUserID)
	if targetRole == "owner" {
		return ErrNotAuthorized
	}

	mute := &models.ChannelMute{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    targetUserID,
		MutedBy:   userID,
		Reason:    req.Reason,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.moderationRepo.CreateMute(ctx, mute); err != nil {
		return err
	}

	s.moderationRepo.LogAction(ctx, &models.ModerationEntry{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    targetUserID,
		Action:    "mute",
		ActorID:   userID,
		Reason:    req.Reason,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *ChannelService) UnmuteMember(ctx context.Context, channelID, targetUserID, userID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	if err := s.moderationRepo.RemoveMute(ctx, channelID, targetUserID); err != nil {
		return err
	}

	s.moderationRepo.LogAction(ctx, &models.ModerationEntry{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    targetUserID,
		Action:    "unmute",
		ActorID:   userID,
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *ChannelService) GetModerationHistory(ctx context.Context, channelID, userID string, limit, offset int) ([]*models.ModerationEntry, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}
	if limit <= 0 {
		limit = 50
	}
	return s.moderationRepo.GetHistory(ctx, channelID, limit, offset)
}

// ── Announcements ──

func (s *ChannelService) CreateAnnouncement(ctx context.Context, channelID, userID string, req *models.CreateAnnouncementRequest) (*models.ChannelAnnouncement, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	ann := &models.ChannelAnnouncement{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		Title:     req.Title,
		Content:   req.Content,
		Priority:  priority,
		AuthorID:  userID,
		IsPinned:  false,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.announcementRepo.Create(ctx, ann); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "channel-events", channelID, "announcement.created", map[string]interface{}{
		"channel_id":      channelID,
		"announcement_id": ann.ID,
	})

	return ann, nil
}

func (s *ChannelService) ListAnnouncements(ctx context.Context, channelID, userID string) ([]*models.ChannelAnnouncement, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.announcementRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) UpdateAnnouncement(ctx context.Context, channelID, announcementID, userID string, req *models.UpdateAnnouncementRequest) (*models.ChannelAnnouncement, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	ann, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil || ann == nil {
		return nil, ErrAnnouncementNotFound
	}

	if req.Title != nil {
		ann.Title = *req.Title
	}
	if req.Content != nil {
		ann.Content = *req.Content
	}
	if req.Priority != nil {
		ann.Priority = *req.Priority
	}
	if req.ExpiresAt != nil {
		ann.ExpiresAt = req.ExpiresAt
	}

	if err := s.announcementRepo.Update(ctx, ann); err != nil {
		return nil, err
	}

	return ann, nil
}

func (s *ChannelService) DeleteAnnouncement(ctx context.Context, channelID, announcementID, userID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	ann, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil || ann == nil {
		return ErrAnnouncementNotFound
	}

	return s.announcementRepo.Delete(ctx, announcementID)
}

func (s *ChannelService) PinAnnouncement(ctx context.Context, channelID, announcementID, userID string) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	ann, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil || ann == nil {
		return ErrAnnouncementNotFound
	}

	return s.announcementRepo.TogglePin(ctx, announcementID, !ann.IsPinned)
}

// ── Sections ──

func (s *ChannelService) CreateSection(ctx context.Context, userID string, req *models.CreateSectionRequest) (*models.ChannelSection, error) {
	channelIDsJSON, _ := json.Marshal(req.ChannelIDs)
	maxPos, _ := s.sectionRepo.GetMaxPosition(ctx, req.WorkspaceID, userID)

	section := &models.ChannelSection{
		ID:          uuid.New().String(),
		WorkspaceID: req.WorkspaceID,
		UserID:      userID,
		Name:        req.Name,
		Position:    maxPos + 1,
		IsCollapsed: false,
		ChannelIDs:  string(channelIDsJSON),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.sectionRepo.Create(ctx, section); err != nil {
		return nil, err
	}

	return section, nil
}

func (s *ChannelService) ListSections(ctx context.Context, workspaceID, userID string) ([]*models.ChannelSection, error) {
	return s.sectionRepo.ListByUser(ctx, workspaceID, userID)
}

func (s *ChannelService) UpdateSection(ctx context.Context, sectionID, userID string, req *models.UpdateSectionRequest) (*models.ChannelSection, error) {
	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil || section == nil {
		return nil, ErrSectionNotFound
	}
	if section.UserID != userID {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		section.Name = *req.Name
	}
	if req.IsCollapsed != nil {
		section.IsCollapsed = *req.IsCollapsed
	}
	if req.ChannelIDs != nil {
		channelIDsJSON, _ := json.Marshal(req.ChannelIDs)
		section.ChannelIDs = string(channelIDsJSON)
	}

	if err := s.sectionRepo.Update(ctx, section); err != nil {
		return nil, err
	}

	return section, nil
}

func (s *ChannelService) DeleteSection(ctx context.Context, sectionID, userID string) error {
	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil || section == nil {
		return ErrSectionNotFound
	}
	if section.UserID != userID {
		return ErrNotAuthorized
	}

	return s.sectionRepo.Delete(ctx, sectionID)
}

// ── Analytics ──

func (s *ChannelService) GetChannelStats(ctx context.Context, channelID, userID string) (*models.ChannelStats, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, channelID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.analyticsRepo.GetChannelStats(ctx, channelID)
}

func (s *ChannelService) GetChannelAnalytics(ctx context.Context, channelID, userID string, days int) ([]models.ChannelActivity, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}
	if days <= 0 {
		days = 30
	}
	return s.analyticsRepo.GetDailyActivity(ctx, channelID, days)
}

// ── Transfer Ownership ──

func (s *ChannelService) TransferOwnership(ctx context.Context, channelID, userID string, req *models.TransferOwnershipRequest) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" {
		return ErrNotAuthorized
	}

	isMember, _ := s.memberRepo.IsMember(ctx, channelID, req.NewOwnerID)
	if !isMember {
		return ErrNotMember
	}

	if err := s.memberRepo.UpdateRole(ctx, channelID, req.NewOwnerID, "owner"); err != nil {
		return err
	}
	if err := s.memberRepo.UpdateRole(ctx, channelID, userID, "admin"); err != nil {
		return err
	}

	s.publishEvent(ctx, "channel-events", channelID, "ownership.transferred", map[string]interface{}{
		"channel_id":   channelID,
		"old_owner":    userID,
		"new_owner":    req.NewOwnerID,
	})

	return nil
}

// ── Bulk Member Operations ──

func (s *ChannelService) BulkAddMembers(ctx context.Context, channelID, userID string, req *models.BulkAddMembersRequest) (int, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return 0, ErrNotAuthorized
	}

	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil || ch == nil {
		return 0, ErrChannelNotFound
	}
	if ch.IsArchived {
		return 0, ErrChannelArchived
	}

	memberRole := req.Role
	if memberRole == "" {
		memberRole = "member"
	}

	added := 0
	for _, uid := range req.UserIDs {
		existing, _ := s.memberRepo.GetByChannelAndUser(ctx, channelID, uid)
		if existing != nil {
			continue
		}
		banned, _ := s.moderationRepo.IsBanned(ctx, channelID, uid)
		if banned {
			continue
		}

		member := &models.ChannelMember{
			ID:            uuid.New().String(),
			ChannelID:     channelID,
			UserID:        uid,
			Role:          memberRole,
			Notifications: "all",
			JoinedAt:      time.Now(),
		}
		if err := s.memberRepo.Create(ctx, member); err == nil {
			added++
		}
	}

	if added > 0 {
		s.invalidateChannel(ctx, channelID)
	}

	return added, nil
}

// ── Clone Channel ──

func (s *ChannelService) CloneChannel(ctx context.Context, channelID, userID string, req *models.CloneChannelRequest) (*models.Channel, error) {
	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil || ch == nil {
		return nil, ErrChannelNotFound
	}

	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	newCh := &models.Channel{
		ID:          uuid.New().String(),
		WorkspaceID: ch.WorkspaceID,
		Name:        req.Name,
		Type:        ch.Type,
		Description: ch.Description,
		CreatedBy:   &userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.channelRepo.Create(ctx, newCh); err != nil {
		return nil, err
	}

	// Add creator as owner
	ownerMember := &models.ChannelMember{
		ID:            uuid.New().String(),
		ChannelID:     newCh.ID,
		UserID:        userID,
		Role:          "owner",
		Notifications: "all",
		JoinedAt:      time.Now(),
	}
	s.memberRepo.Create(ctx, ownerMember)

	// Optionally clone members
	if req.IncludeMembers {
		members, _ := s.memberRepo.ListByChannel(ctx, channelID)
		for _, m := range members {
			if m.UserID == userID {
				continue
			}
			newMember := &models.ChannelMember{
				ID:            uuid.New().String(),
				ChannelID:     newCh.ID,
				UserID:        m.UserID,
				Role:          m.Role,
				Notifications: m.Notifications,
				JoinedAt:      time.Now(),
			}
			s.memberRepo.Create(ctx, newMember)
		}
	}

	// Optionally clone pins
	if req.IncludePins {
		pins, _ := s.pinRepo.ListByChannel(ctx, channelID)
		for _, p := range pins {
			newPin := &models.ChannelPin{
				ID:        uuid.New().String(),
				ChannelID: newCh.ID,
				MessageID: p.MessageID,
				PinnedBy:  userID,
				PinnedAt:  time.Now(),
			}
			s.pinRepo.Create(ctx, newPin)
		}
	}

	return newCh, nil
}

// ── Threads ──

func (s *ChannelService) CreateThread(ctx context.Context, channelID, userID string, req *models.CreateThreadRequest) (*models.ChannelThread, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role == "" {
		return nil, ErrNotMember
	}

	thread := &models.ChannelThread{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		MessageID: req.MessageID,
		Title:     req.Title,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.threadRepo.Create(ctx, thread); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "channel-events", channelID, "thread.created", map[string]interface{}{
		"thread":     thread,
		"channel_id": channelID,
	})

	return thread, nil
}

func (s *ChannelService) GetThread(ctx context.Context, threadID string) (*models.ChannelThread, error) {
	thread, err := s.threadRepo.GetByID(ctx, threadID)
	if err != nil || thread == nil {
		return nil, ErrThreadNotFound
	}
	return thread, nil
}

func (s *ChannelService) ListThreads(ctx context.Context, channelID string, limit, offset int) ([]*models.ChannelThread, error) {
	return s.threadRepo.ListByChannel(ctx, channelID, limit, offset)
}

func (s *ChannelService) DeleteThread(ctx context.Context, channelID, threadID, userID string) error {
	thread, err := s.threadRepo.GetByID(ctx, threadID)
	if err != nil || thread == nil {
		return ErrThreadNotFound
	}

	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" && thread.CreatedBy != userID {
		return ErrNotAuthorized
	}

	return s.threadRepo.Delete(ctx, threadID)
}

func (s *ChannelService) CreateReply(ctx context.Context, threadID, userID string, req *models.CreateReplyRequest) (*models.ThreadReply, error) {
	thread, err := s.threadRepo.GetByID(ctx, threadID)
	if err != nil || thread == nil {
		return nil, ErrThreadNotFound
	}

	if thread.IsLocked {
		return nil, ErrNotAuthorized
	}

	role, _ := s.memberRepo.GetRole(ctx, thread.ChannelID, userID)
	if role == "" {
		return nil, ErrNotMember
	}

	reply := &models.ThreadReply{
		ID:        uuid.New().String(),
		ThreadID:  threadID,
		UserID:    userID,
		Content:   req.Content,
		ParentID:  req.ParentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.threadRepo.CreateReply(ctx, reply); err != nil {
		return nil, err
	}

	s.threadRepo.IncrementReplyCount(ctx, threadID)

	return reply, nil
}

func (s *ChannelService) ListReplies(ctx context.Context, threadID string, limit, offset int) ([]*models.ThreadReply, error) {
	thread, err := s.threadRepo.GetByID(ctx, threadID)
	if err != nil || thread == nil {
		return nil, ErrThreadNotFound
	}
	return s.threadRepo.ListReplies(ctx, threadID, limit, offset)
}

func (s *ChannelService) UpdateReply(ctx context.Context, replyID, userID string, req *models.UpdateReplyRequest) (*models.ThreadReply, error) {
	reply, err := s.threadRepo.GetReplyByID(ctx, replyID)
	if err != nil || reply == nil {
		return nil, ErrReplyNotFound
	}

	if reply.UserID != userID {
		return nil, ErrNotAuthorized
	}

	reply.Content = req.Content
	reply.UpdatedAt = time.Now()

	if err := s.threadRepo.UpdateReply(ctx, reply); err != nil {
		return nil, err
	}

	return reply, nil
}

func (s *ChannelService) DeleteReply(ctx context.Context, channelID, replyID, userID string) error {
	reply, err := s.threadRepo.GetReplyByID(ctx, replyID)
	if err != nil || reply == nil {
		return ErrReplyNotFound
	}

	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" && reply.UserID != userID {
		return ErrNotAuthorized
	}

	if err := s.threadRepo.DeleteReply(ctx, replyID); err != nil {
		return err
	}

	thread, _ := s.threadRepo.GetByID(ctx, reply.ThreadID)
	if thread != nil {
		s.threadRepo.DecrementReplyCount(ctx, reply.ThreadID)
	}

	return nil
}

func (s *ChannelService) FollowThread(ctx context.Context, threadID, userID string) error {
	thread, err := s.threadRepo.GetByID(ctx, threadID)
	if err != nil || thread == nil {
		return ErrThreadNotFound
	}

	following, _ := s.threadRepo.IsFollowing(ctx, threadID, userID)
	if following {
		return nil
	}

	follower := &models.ThreadFollower{
		ID:        uuid.New().String(),
		ThreadID:  threadID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	return s.threadRepo.AddFollower(ctx, follower)
}

func (s *ChannelService) UnfollowThread(ctx context.Context, threadID, userID string) error {
	return s.threadRepo.RemoveFollower(ctx, threadID, userID)
}

// ── Channel Settings ──

func (s *ChannelService) GetSettings(ctx context.Context, channelID, userID string) (*models.ChannelSetting, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role == "" {
		return nil, ErrNotMember
	}

	settings, err := s.settingsRepo.Get(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return &models.ChannelSetting{ChannelID: channelID}, nil
	}
	return settings, nil
}

func (s *ChannelService) UpdateSettings(ctx context.Context, channelID, userID string, req *models.UpdateSettingsRequest) (*models.ChannelSetting, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	setting := &models.ChannelSetting{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.SlowModeInterval != nil {
		setting.SlowModeInterval = *req.SlowModeInterval
	}
	if req.MaxPins != nil {
		setting.MaxPins = *req.MaxPins
	}
	if req.MaxBookmarks != nil {
		setting.MaxBookmarks = *req.MaxBookmarks
	}
	if req.AllowThreads != nil {
		setting.AllowThreads = *req.AllowThreads
	}
	if req.AllowReactions != nil {
		setting.AllowReactions = *req.AllowReactions
	}
	if req.AllowInvites != nil {
		setting.AllowInvites = *req.AllowInvites
	}
	if req.AutoArchiveDays != nil {
		setting.AutoArchiveDays = *req.AutoArchiveDays
	}
	if req.DefaultNotification != nil {
		setting.DefaultNotification = *req.DefaultNotification
	}
	if req.CustomEmoji != nil {
		setting.CustomEmoji = *req.CustomEmoji
	}
	if req.LinkPreviews != nil {
		setting.LinkPreviews = *req.LinkPreviews
	}
	if req.MemberLimit != nil {
		setting.MemberLimit = *req.MemberLimit
	}

	if err := s.settingsRepo.Upsert(ctx, setting); err != nil {
		return nil, err
	}

	result, _ := s.settingsRepo.Get(ctx, channelID)
	if result == nil {
		return setting, nil
	}

	s.activityLogRepo.Create(ctx, &models.ChannelActivityLog{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    userID,
		Action:    "settings_updated",
		CreatedAt: time.Now(),
	})

	return result, nil
}

// ── Starred Channels ──

func (s *ChannelService) StarChannel(ctx context.Context, channelID, userID string) (*models.StarredChannel, error) {
	starred, _ := s.starredRepo.IsStarred(ctx, userID, channelID)
	if starred {
		return nil, ErrAlreadyStarred
	}

	maxPos, _ := s.starredRepo.GetMaxPosition(ctx, userID)

	star := &models.StarredChannel{
		ID:        uuid.New().String(),
		UserID:    userID,
		ChannelID: channelID,
		Position:  maxPos + 1,
		CreatedAt: time.Now(),
	}

	if err := s.starredRepo.Star(ctx, star); err != nil {
		return nil, err
	}

	return star, nil
}

func (s *ChannelService) UnstarChannel(ctx context.Context, channelID, userID string) error {
	starred, _ := s.starredRepo.IsStarred(ctx, userID, channelID)
	if !starred {
		return ErrNotStarred
	}
	return s.starredRepo.Unstar(ctx, userID, channelID)
}

func (s *ChannelService) ListStarredChannels(ctx context.Context, userID string) ([]*models.StarredChannel, error) {
	return s.starredRepo.ListByUser(ctx, userID)
}

// ── Read Receipts ──

func (s *ChannelService) MarkRead(ctx context.Context, channelID, userID string, req *models.MarkReadRequest) error {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role == "" {
		return ErrNotMember
	}

	receipt := &models.ReadReceipt{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    userID,
		MessageID: req.MessageID,
		ReadAt:    time.Now(),
	}

	return s.readReceiptRepo.Upsert(ctx, receipt)
}

func (s *ChannelService) GetReadReceipts(ctx context.Context, channelID, messageID string) ([]*models.ReadReceipt, error) {
	return s.readReceiptRepo.ListByMessage(ctx, channelID, messageID)
}

func (s *ChannelService) GetReadCount(ctx context.Context, channelID, messageID string) (int, error) {
	return s.readReceiptRepo.GetReadCount(ctx, channelID, messageID)
}

// ── Scheduled Messages ──

func (s *ChannelService) CreateScheduledMessage(ctx context.Context, channelID, userID string, req *models.CreateScheduledMessageRequest) (*models.ScheduledMessage, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role == "" {
		return nil, ErrNotMember
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled_at format: %w", err)
	}

	msg := &models.ScheduledMessage{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		UserID:      userID,
		Content:     req.Content,
		ScheduledAt: scheduledAt,
		Status:      "pending",
		ThreadID:    req.ThreadID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.scheduledMessageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *ChannelService) GetScheduledMessage(ctx context.Context, messageID string) (*models.ScheduledMessage, error) {
	msg, err := s.scheduledMessageRepo.GetByID(ctx, messageID)
	if err != nil || msg == nil {
		return nil, ErrScheduledMessageNotFound
	}
	return msg, nil
}

func (s *ChannelService) ListScheduledMessages(ctx context.Context, channelID, userID string) ([]*models.ScheduledMessage, error) {
	return s.scheduledMessageRepo.ListByChannel(ctx, channelID, userID)
}

func (s *ChannelService) ListMyScheduledMessages(ctx context.Context, userID string) ([]*models.ScheduledMessage, error) {
	return s.scheduledMessageRepo.ListByUser(ctx, userID)
}

func (s *ChannelService) UpdateScheduledMessage(ctx context.Context, messageID, userID string, req *models.UpdateScheduledMessageRequest) (*models.ScheduledMessage, error) {
	msg, err := s.scheduledMessageRepo.GetByID(ctx, messageID)
	if err != nil || msg == nil {
		return nil, ErrScheduledMessageNotFound
	}

	if msg.UserID != userID {
		return nil, ErrNotAuthorized
	}

	if msg.Status != "pending" {
		return nil, fmt.Errorf("cannot update a %s message", msg.Status)
	}

	if req.Content != nil {
		msg.Content = *req.Content
	}
	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			return nil, fmt.Errorf("invalid scheduled_at format: %w", err)
		}
		msg.ScheduledAt = t
	}
	msg.UpdatedAt = time.Now()

	if err := s.scheduledMessageRepo.Update(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *ChannelService) DeleteScheduledMessage(ctx context.Context, messageID, userID string) error {
	msg, err := s.scheduledMessageRepo.GetByID(ctx, messageID)
	if err != nil || msg == nil {
		return ErrScheduledMessageNotFound
	}

	if msg.UserID != userID {
		return ErrNotAuthorized
	}

	return s.scheduledMessageRepo.Delete(ctx, messageID)
}

// ── Activity Log ──

func (s *ChannelService) GetActivityLog(ctx context.Context, channelID, userID string, limit, offset int) ([]*models.ChannelActivityLog, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}
	return s.activityLogRepo.ListByChannel(ctx, channelID, limit, offset)
}

func (s *ChannelService) GetUserActivityLog(ctx context.Context, userID string, limit, offset int) ([]*models.ChannelActivityLog, error) {
	return s.activityLogRepo.ListByUser(ctx, userID, limit, offset)
}

// ── Channel Templates ──

func (s *ChannelService) CreateTemplate(ctx context.Context, userID string, req *models.CreateTemplateRequest) (*models.ChannelTemplate, error) {
	tmpl := &models.ChannelTemplate{
		ID:          uuid.New().String(),
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Topic:       req.Topic,
		Settings:    req.Settings,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if tmpl.Type == "" {
		tmpl.Type = "public"
	}

	if err := s.templateRepo.Create(ctx, tmpl); err != nil {
		return nil, err
	}

	return tmpl, nil
}

func (s *ChannelService) GetTemplate(ctx context.Context, templateID string) (*models.ChannelTemplate, error) {
	tmpl, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil || tmpl == nil {
		return nil, ErrTemplateNotFound
	}
	return tmpl, nil
}

func (s *ChannelService) ListTemplates(ctx context.Context, workspaceID string) ([]*models.ChannelTemplate, error) {
	return s.templateRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *ChannelService) UpdateTemplate(ctx context.Context, templateID, userID string, req *models.UpdateTemplateRequest) (*models.ChannelTemplate, error) {
	tmpl, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil || tmpl == nil {
		return nil, ErrTemplateNotFound
	}

	if tmpl.CreatedBy != userID {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		tmpl.Name = *req.Name
	}
	if req.Description != nil {
		tmpl.Description = req.Description
	}
	if req.Type != nil {
		tmpl.Type = *req.Type
	}
	if req.Topic != nil {
		tmpl.Topic = req.Topic
	}
	if req.Settings != nil {
		tmpl.Settings = req.Settings
	}
	tmpl.UpdatedAt = time.Now()

	if err := s.templateRepo.Update(ctx, tmpl); err != nil {
		return nil, err
	}

	return tmpl, nil
}

func (s *ChannelService) DeleteTemplate(ctx context.Context, templateID, userID string) error {
	tmpl, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil || tmpl == nil {
		return ErrTemplateNotFound
	}

	if tmpl.CreatedBy != userID {
		return ErrNotAuthorized
	}

	return s.templateRepo.Delete(ctx, templateID)
}

func (s *ChannelService) ApplyTemplate(ctx context.Context, templateID, userID string, req *models.ApplyTemplateRequest) (*models.Channel, error) {
	tmpl, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil || tmpl == nil {
		return nil, ErrTemplateNotFound
	}

	createReq := &models.CreateChannelRequest{
		WorkspaceID: tmpl.WorkspaceID,
		Name:        req.ChannelName,
		Type:        tmpl.Type,
	}

	ch, err := s.CreateChannel(ctx, userID, createReq)
	if err != nil {
		return nil, err
	}

	if tmpl.Topic != nil {
		updateReq := &models.UpdateChannelRequest{Topic: tmpl.Topic}
		s.UpdateChannel(ctx, ch.ID, userID, updateReq)
	}

	return ch, nil
}

// ── Voice Channels ──

func (s *ChannelService) JoinVoice(ctx context.Context, channelID, userID string) (*models.VoiceChannelState, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role == "" {
		return nil, ErrNotMember
	}

	existing, _ := s.voiceRepo.GetState(ctx, channelID, userID)
	if existing != nil {
		return nil, ErrAlreadyInVoice
	}

	state := &models.VoiceChannelState{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    userID,
		JoinedAt:  time.Now(),
	}

	if err := s.voiceRepo.Join(ctx, state); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "channel-events", channelID, "voice.joined", map[string]interface{}{
		"channel_id": channelID,
		"user_id":    userID,
	})

	return state, nil
}

func (s *ChannelService) LeaveVoice(ctx context.Context, channelID, userID string) error {
	existing, _ := s.voiceRepo.GetState(ctx, channelID, userID)
	if existing == nil {
		return ErrNotInVoice
	}

	if err := s.voiceRepo.Leave(ctx, channelID, userID); err != nil {
		return err
	}

	s.publishEvent(ctx, "channel-events", channelID, "voice.left", map[string]interface{}{
		"channel_id": channelID,
		"user_id":    userID,
	})

	return nil
}

func (s *ChannelService) UpdateVoiceState(ctx context.Context, channelID, userID string, req *models.UpdateVoiceStateRequest) error {
	existing, _ := s.voiceRepo.GetState(ctx, channelID, userID)
	if existing == nil {
		return ErrNotInVoice
	}

	return s.voiceRepo.UpdateState(ctx, channelID, userID, req)
}

func (s *ChannelService) ListVoiceParticipants(ctx context.Context, channelID string) ([]*models.VoiceChannelState, error) {
	return s.voiceRepo.ListParticipants(ctx, channelID)
}

func (s *ChannelService) CountVoiceParticipants(ctx context.Context, channelID string) (int, error) {
	return s.voiceRepo.CountParticipants(ctx, channelID)
}

// ── Channel Followers ──

func (s *ChannelService) FollowChannel(ctx context.Context, channelID, userID string) (*models.ChannelFollower, error) {
	following, _ := s.followerRepo.IsFollowing(ctx, channelID, userID)
	if following {
		return nil, ErrAlreadyFollowing
	}

	follower := &models.ChannelFollower{
		ID:         uuid.New().String(),
		ChannelID:  channelID,
		UserID:     userID,
		FollowedAt: time.Now(),
	}

	if err := s.followerRepo.Follow(ctx, follower); err != nil {
		return nil, err
	}

	return follower, nil
}

func (s *ChannelService) UnfollowChannel(ctx context.Context, channelID, userID string) error {
	following, _ := s.followerRepo.IsFollowing(ctx, channelID, userID)
	if !following {
		return ErrNotFollowing
	}
	return s.followerRepo.Unfollow(ctx, channelID, userID)
}

func (s *ChannelService) ListChannelFollowers(ctx context.Context, channelID string, limit, offset int) ([]*models.ChannelFollower, error) {
	return s.followerRepo.ListByChannel(ctx, channelID, limit, offset)
}

func (s *ChannelService) ListFollowedChannels(ctx context.Context, userID string) ([]*models.ChannelFollower, error) {
	return s.followerRepo.ListByUser(ctx, userID)
}

func (s *ChannelService) CountChannelFollowers(ctx context.Context, channelID string) (int, error) {
	return s.followerRepo.CountByChannel(ctx, channelID)
}

// ── Bulk Operations Extended ──

func (s *ChannelService) BulkDeleteMembers(ctx context.Context, channelID, userID string, req *models.BulkDeleteMembersRequest) (*models.BulkActionResult, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	result := &models.BulkActionResult{}
	for _, uid := range req.UserIDs {
		if uid == userID {
			result.Failed++
			result.Errors = append(result.Errors, "cannot remove yourself")
			continue
		}
		if err := s.memberRepo.Remove(ctx, channelID, uid); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Successful++
		}
	}

	if result.Successful > 0 {
		s.invalidateChannel(ctx, channelID)
	}

	return result, nil
}

func (s *ChannelService) BulkUpdateRoles(ctx context.Context, channelID, userID string, req *models.BulkUpdateRoleRequest) (*models.BulkActionResult, error) {
	role, _ := s.memberRepo.GetRole(ctx, channelID, userID)
	if role != "owner" {
		return nil, ErrNotAuthorized
	}

	result := &models.BulkActionResult{}
	for _, uid := range req.UserIDs {
		if uid == userID {
			result.Failed++
			result.Errors = append(result.Errors, "cannot change own role")
			continue
		}
		if err := s.memberRepo.UpdateRole(ctx, channelID, uid, req.Role); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Successful++
		}
	}

	return result, nil
}

// ── Redis Cache Helpers ──

func (s *ChannelService) cacheChannel(ctx context.Context, id string, ch *models.Channel) {
	if s.redis == nil {
		return
	}
	data, err := json.Marshal(ch)
	if err != nil {
		return
	}
	key := fmt.Sprintf(cacheKeyChannel, id)
	s.redis.Set(ctx, key, data, cacheTTL)
}

func (s *ChannelService) getCachedChannel(ctx context.Context, id string) (*models.Channel, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("no redis")
	}
	key := fmt.Sprintf(cacheKeyChannel, id)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var ch models.Channel
	if err := json.Unmarshal(data, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

func (s *ChannelService) invalidateChannel(ctx context.Context, channelID string) {
	if s.redis == nil {
		return
	}
	keys := []string{
		fmt.Sprintf(cacheKeyChannel, channelID),
		fmt.Sprintf(cacheKeyMembers, channelID),
	}
	s.redis.Del(ctx, keys...)
}

// ── Kafka Event Helper ──

func (s *ChannelService) publishEvent(ctx context.Context, topic, key, eventType string, data map[string]interface{}) {
	if s.kafka == nil {
		return
	}
	data["type"] = eventType
	data["timestamp"] = time.Now()
	if err := s.kafka.Publish(ctx, topic, key, data); err != nil {
		s.logger.WithError(err).WithField("event_type", eventType).Warn("Failed to publish event")
	}
}
