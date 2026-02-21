package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/quckapp/channel-service/internal/models"
	"github.com/quckapp/channel-service/internal/repository"
	"github.com/sirupsen/logrus"
)

var (
	ErrPollNotFound             = errors.New("poll not found")
	ErrAlreadyVoted             = errors.New("user has already voted on this poll")
	ErrPollClosed               = errors.New("poll is closed")
	ErrScheduledMessageNotFound = errors.New("scheduled message not found")
	ErrChannelLinkNotFound      = errors.New("channel link not found")
	ErrTabNotFound              = errors.New("tab not found")
	ErrAlreadyFollowing         = errors.New("already following this channel")
	ErrChannelTemplateNotFound  = errors.New("channel template not found")
	ErrNotFollowing             = errors.New("not following this channel")
	ErrScheduledTimeInPast      = errors.New("scheduled time must be in the future")
)

type ChannelService struct {
	pollRepo             *repository.PollRepository
	scheduledMessageRepo *repository.ScheduledMessageRepository
	channelLinkRepo      *repository.ChannelLinkRepository
	tabRepo              *repository.TabRepository
	followerRepo         *repository.FollowerRepository
	templateRepo         *repository.TemplateRepository
	logger               *logrus.Logger
}

func NewChannelService(
	pollRepo *repository.PollRepository,
	scheduledMessageRepo *repository.ScheduledMessageRepository,
	channelLinkRepo *repository.ChannelLinkRepository,
	tabRepo *repository.TabRepository,
	followerRepo *repository.FollowerRepository,
	templateRepo *repository.TemplateRepository,
	logger *logrus.Logger,
) *ChannelService {
	return &ChannelService{
		pollRepo:             pollRepo,
		scheduledMessageRepo: scheduledMessageRepo,
		channelLinkRepo:      channelLinkRepo,
		tabRepo:              tabRepo,
		followerRepo:         followerRepo,
		templateRepo:         templateRepo,
		logger:               logger,
	}
}

// ── Polls ──

func (s *ChannelService) CreatePoll(ctx context.Context, channelID, userID string, req *models.CreatePollRequest) (*models.PollWithOptions, error) {
	now := time.Now()
	poll := &models.ChannelPoll{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		CreatedBy:   userID,
		Question:    req.Question,
		IsAnonymous: req.IsAnonymous,
		MultiChoice: req.MultiChoice,
		IsClosed:    false,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.pollRepo.Create(ctx, poll); err != nil {
		return nil, err
	}

	var options []models.PollOption
	for i, optText := range req.Options {
		option := &models.PollOption{
			ID:       uuid.New().String(),
			PollID:   poll.ID,
			Text:     optText,
			Position: i,
		}
		if err := s.pollRepo.CreateOption(ctx, option); err != nil {
			return nil, err
		}
		options = append(options, *option)
	}

	return &models.PollWithOptions{
		ChannelPoll: *poll,
		Options:     options,
	}, nil
}

func (s *ChannelService) GetPoll(ctx context.Context, pollID string) (*models.PollWithOptions, error) {
	poll, err := s.pollRepo.GetByID(ctx, pollID)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, ErrPollNotFound
	}

	options, err := s.pollRepo.GetOptions(ctx, pollID)
	if err != nil {
		return nil, err
	}

	return &models.PollWithOptions{
		ChannelPoll: *poll,
		Options:     options,
	}, nil
}

func (s *ChannelService) ListPolls(ctx context.Context, channelID string) ([]*models.ChannelPoll, error) {
	return s.pollRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) VotePoll(ctx context.Context, pollID, userID string, req *models.VotePollRequest) error {
	poll, err := s.pollRepo.GetByID(ctx, pollID)
	if err != nil {
		return err
	}
	if poll == nil {
		return ErrPollNotFound
	}
	if poll.IsClosed {
		return ErrPollClosed
	}

	hasVoted, err := s.pollRepo.HasVoted(ctx, pollID, userID)
	if err != nil {
		return err
	}
	if hasVoted {
		return ErrAlreadyVoted
	}

	now := time.Now()
	for _, optionID := range req.OptionIDs {
		vote := &models.PollVote{
			ID:       uuid.New().String(),
			PollID:   pollID,
			OptionID: optionID,
			UserID:   userID,
			VotedAt:  now,
		}
		if err := s.pollRepo.CreateVote(ctx, vote); err != nil {
			return err
		}
	}

	return nil
}

func (s *ChannelService) ClosePoll(ctx context.Context, pollID, userID string) error {
	poll, err := s.pollRepo.GetByID(ctx, pollID)
	if err != nil {
		return err
	}
	if poll == nil {
		return ErrPollNotFound
	}
	if poll.IsClosed {
		return ErrPollClosed
	}

	now := time.Now()
	return s.pollRepo.ClosePoll(ctx, pollID, now)
}

func (s *ChannelService) GetPollResults(ctx context.Context, pollID string) (*models.PollResults, error) {
	poll, err := s.pollRepo.GetByID(ctx, pollID)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, ErrPollNotFound
	}

	results, err := s.pollRepo.GetResults(ctx, pollID)
	if err != nil {
		return nil, err
	}

	totalVotes, err := s.pollRepo.GetTotalVotes(ctx, pollID)
	if err != nil {
		return nil, err
	}

	return &models.PollResults{
		Poll:       *poll,
		Results:    results,
		TotalVotes: totalVotes,
	}, nil
}

// ── Scheduled Messages ──

func (s *ChannelService) ScheduleMessage(ctx context.Context, channelID, userID string, req *models.CreateScheduledMessageRequest) (*models.ScheduledMessage, error) {
	if req.ScheduledAt.Before(time.Now()) {
		return nil, ErrScheduledTimeInPast
	}

	now := time.Now()
	msg := &models.ScheduledMessage{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		UserID:      userID,
		Content:     req.Content,
		ScheduledAt: req.ScheduledAt,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.scheduledMessageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *ChannelService) ListScheduledMessages(ctx context.Context, channelID string) ([]*models.ScheduledMessage, error) {
	return s.scheduledMessageRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) GetScheduledMessage(ctx context.Context, messageID string) (*models.ScheduledMessage, error) {
	msg, err := s.scheduledMessageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, ErrScheduledMessageNotFound
	}
	return msg, nil
}

func (s *ChannelService) UpdateScheduledMessage(ctx context.Context, messageID, userID string, req *models.UpdateScheduledMessageRequest) (*models.ScheduledMessage, error) {
	msg, err := s.scheduledMessageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, ErrScheduledMessageNotFound
	}

	if req.Content != nil {
		msg.Content = *req.Content
	}
	if req.ScheduledAt != nil {
		if req.ScheduledAt.Before(time.Now()) {
			return nil, ErrScheduledTimeInPast
		}
		msg.ScheduledAt = *req.ScheduledAt
	}

	if err := s.scheduledMessageRepo.Update(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *ChannelService) CancelScheduledMessage(ctx context.Context, messageID, userID string) error {
	msg, err := s.scheduledMessageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg == nil {
		return ErrScheduledMessageNotFound
	}

	return s.scheduledMessageRepo.Cancel(ctx, messageID)
}

// ── Channel Links ──

func (s *ChannelService) CreateChannelLink(ctx context.Context, channelID, userID string, req *models.CreateChannelLinkRequest) (*models.ChannelLink, error) {
	now := time.Now()
	link := &models.ChannelLink{
		ID:              uuid.New().String(),
		SourceChannelID: channelID,
		TargetChannelID: req.TargetChannelID,
		CreatedBy:       userID,
		LinkType:        req.LinkType,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.channelLinkRepo.Create(ctx, link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *ChannelService) ListChannelLinks(ctx context.Context, channelID string) ([]*models.ChannelLink, error) {
	return s.channelLinkRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) GetChannelLink(ctx context.Context, linkID string) (*models.ChannelLink, error) {
	link, err := s.channelLinkRepo.GetByID(ctx, linkID)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrChannelLinkNotFound
	}
	return link, nil
}

func (s *ChannelService) DeleteChannelLink(ctx context.Context, linkID, userID string) error {
	link, err := s.channelLinkRepo.GetByID(ctx, linkID)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrChannelLinkNotFound
	}

	return s.channelLinkRepo.Delete(ctx, linkID)
}

// ── Channel Tabs ──

func (s *ChannelService) AddTab(ctx context.Context, channelID, userID string, req *models.CreateTabRequest) (*models.ChannelTab, error) {
	maxPos, err := s.tabRepo.GetMaxPosition(ctx, channelID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	tab := &models.ChannelTab{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		Name:      req.Name,
		TabType:   req.TabType,
		Config:    req.Config,
		Position:  maxPos + 1,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.tabRepo.Create(ctx, tab); err != nil {
		return nil, err
	}

	return tab, nil
}

func (s *ChannelService) ListTabs(ctx context.Context, channelID string) ([]*models.ChannelTab, error) {
	return s.tabRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) UpdateTab(ctx context.Context, tabID, userID string, req *models.UpdateTabRequest) (*models.ChannelTab, error) {
	tab, err := s.tabRepo.GetByID(ctx, tabID)
	if err != nil {
		return nil, err
	}
	if tab == nil {
		return nil, ErrTabNotFound
	}

	if req.Name != nil {
		tab.Name = *req.Name
	}
	if req.Config != nil {
		tab.Config = req.Config
	}

	if err := s.tabRepo.Update(ctx, tab); err != nil {
		return nil, err
	}

	return tab, nil
}

func (s *ChannelService) RemoveTab(ctx context.Context, tabID, userID string) error {
	tab, err := s.tabRepo.GetByID(ctx, tabID)
	if err != nil {
		return err
	}
	if tab == nil {
		return ErrTabNotFound
	}

	return s.tabRepo.Delete(ctx, tabID)
}

func (s *ChannelService) ReorderTabs(ctx context.Context, channelID, userID string, req *models.ReorderTabsRequest) error {
	return s.tabRepo.UpdatePositions(ctx, channelID, req.TabIDs)
}

// ── Channel Followers ──

func (s *ChannelService) FollowChannel(ctx context.Context, channelID, userID string) (*models.ChannelFollower, error) {
	existing, err := s.followerRepo.GetByChannelAndUser(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyFollowing
	}

	follower := &models.ChannelFollower{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := s.followerRepo.Create(ctx, follower); err != nil {
		return nil, err
	}

	return follower, nil
}

func (s *ChannelService) UnfollowChannel(ctx context.Context, channelID, userID string) error {
	isFollowing, err := s.followerRepo.IsFollowing(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if !isFollowing {
		return ErrNotFollowing
	}

	return s.followerRepo.Delete(ctx, channelID, userID)
}

func (s *ChannelService) ListFollowers(ctx context.Context, channelID string) ([]*models.ChannelFollower, error) {
	return s.followerRepo.ListByChannel(ctx, channelID)
}

func (s *ChannelService) CheckFollowing(ctx context.Context, channelID, userID string) (bool, error) {
	return s.followerRepo.IsFollowing(ctx, channelID, userID)
}

// ── Channel Templates ──

func (s *ChannelService) CreateTemplate(ctx context.Context, channelID, userID string, req *models.CreateTemplateRequest) (*models.ChannelTemplate, error) {
	now := time.Now()
	tmpl := &models.ChannelTemplate{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
		ChannelType: "public",
		IsPublic:    req.IsPublic,
		UseCount:    0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.templateRepo.Create(ctx, tmpl); err != nil {
		return nil, err
	}

	return tmpl, nil
}

func (s *ChannelService) ListTemplates(ctx context.Context) ([]*models.ChannelTemplate, error) {
	return s.templateRepo.List(ctx)
}

func (s *ChannelService) ApplyTemplate(ctx context.Context, templateID, userID string, req *models.ApplyTemplateRequest) (*models.ChannelTemplate, error) {
	tmpl, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if tmpl == nil {
		return nil, ErrChannelTemplateNotFound
	}

	if err := s.templateRepo.IncrementUseCount(ctx, templateID); err != nil {
		return nil, err
	}

	return tmpl, nil
}

func (s *ChannelService) DeleteTemplate(ctx context.Context, templateID, userID string) error {
	tmpl, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return err
	}
	if tmpl == nil {
		return ErrChannelTemplateNotFound
	}

	return s.templateRepo.Delete(ctx, templateID)
}
