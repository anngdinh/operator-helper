package notify

type AlertConfig struct {
	Title    string            `mapstructure:"title"`
	Metadata map[string]string `mapstructure:"metadata"`
	MSTeams  *MSTeamsConfig    `mapstructure:"msteams"`
	Telegram *TelegramConfig   `mapstructure:"telegram"`
}

type MSTeamsConfig struct {
	WebhookURL string `mapstructure:"webhookUrl"`
}

type TelegramConfig struct {
	BotToken string `mapstructure:"botToken"`
	ChatID   int64  `mapstructure:"chatId"`
	ProxyURL string `mapstructure:"proxyUrl"`
}
