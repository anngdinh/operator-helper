package notify

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/msteams"
	"github.com/nikoksr/notify/service/telegram"

	"github.com/anngdinh/operator-helper/notify/workerpool"
)

// Notifier interface for sending notifications
type Notifier interface {
	// Send sends a notification with the given status, fields, and content
	Send(status Status, fields map[string]string, content string)
}

type notifierImpl struct {
	ctx           context.Context
	sendAlertPool *workerpool.Pool
	notifier      *notify.Notify
	msgBuilder    *MessageBuilder
	logger        logr.Logger
}

// NewNotifier creates a new Notifier with fixed title and base metadata from config
func NewNotifier(ctx context.Context, alertConfig *AlertConfig, logger logr.Logger) (Notifier, error) {
	sendAlertPool := workerpool.NewPool(2, workerpool.BackoffConfig{
		InitialDelay: 10 * time.Second,
		MaxRetries:   3,
		Factor:       2.0,
	})
	sendAlertPool.Start()

	notifierServices := []notify.Notifier{}
	if alertConfig.MSTeams != nil && alertConfig.MSTeams.WebhookURL != "" {
		teamsService := msteams.New()
		teamsService.AddReceivers(alertConfig.MSTeams.WebhookURL)
		notifierServices = append(notifierServices, teamsService)
	}
	if alertConfig.Telegram != nil && alertConfig.Telegram.BotToken != "" && alertConfig.Telegram.ChatID != 0 {
		tgService := &telegram.Telegram{}
		var _err error
		if alertConfig.Telegram.ProxyURL != "" {
			proxyURL, _err := url.Parse(alertConfig.Telegram.ProxyURL)
			if _err != nil {
				return nil, fmt.Errorf("failed to parse Telegram proxy URL: %s, err: %s", alertConfig.Telegram.ProxyURL, _err)
			}
			client, _err := tgbotapi.NewBotAPIWithClient(alertConfig.Telegram.BotToken, &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
			})
			if _err == nil {
				tgService.SetClient(client)
			} else {
				logger.Error(_err, "failed to create Telegram service with proxy")
			}
		} else {
			tgService, _err = telegram.New(alertConfig.Telegram.BotToken)
			if _err != nil {
				logger.Error(_err, "failed to create Telegram service")
			}
		}

		// ignore if err
		if _err == nil {
			tgService.SetParseMode(telegram.ModeMarkdown)
			tgService.AddReceivers(alertConfig.Telegram.ChatID)
			notifierServices = append(notifierServices, tgService)
		}
	}
	notifier := notify.New()
	notifier.UseServices(notifierServices...)

	return &notifierImpl{
		ctx:           ctx,
		notifier:      notifier,
		sendAlertPool: sendAlertPool,
		msgBuilder:    NewMessageBuilder(alertConfig.Title, alertConfig.Metadata),
		logger:        logger,
	}, nil
}

// Send creates and sends a notification with the given status, fields, and content
func (s *notifierImpl) Send(status Status, fields map[string]string, content string) {
	notification := s.msgBuilder.NewNotification(status)

	// Auto-add timestamp
	notification.WithField("timestamp", time.Now().In(time.FixedZone("GMT+7", 7*60*60)).Format("2006-01-02 15:04:05"))

	for k, v := range fields {
		notification.WithField(k, v)
	}
	if content != "" {
		notification.WithContent(content)
	}

	title := notification.GetTitle()
	body := notification.GetBodyTelegram()

	s.sendAlertPool.AddTask(
		workerpool.NewTask(func() error {
			err := s.notifier.Send(s.ctx, title, body)
			if err != nil {
				s.logger.Error(err, "failed to send notification")
			}
			return err
		}),
	)
}
