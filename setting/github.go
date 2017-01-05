package setting

type GithubSetting struct {
	BotName    string `toml:"botname"`
	Token      string `toml:"api_token"`
	HookSecret string `toml:"webhook_secret"`
}
