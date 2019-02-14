package setting

type GithubSetting struct {
	BotName      string   `toml:"botname"`
	Token        string   `toml:"api_token"`
	HookSecret   string   `toml:"webhook_secret"`
	Repositories []string `toml:"accepted_repositoies"`
	BaseURL      string   `toml:"base_url"`
	UploadURL    string   `toml:"upload_url"`

	acceptedRepos map[string]bool
}

func initGithubSetting(g *GithubSetting) {
	list := g.Repositories
	g.Repositories = nil

	if (list == nil) || (len(list) == 0) {
		g.acceptedRepos = nil
		return
	}

	m := make(map[string]bool)
	for _, r := range list {
		m[r] = true
	}

	g.acceptedRepos = m
	return
}

func (g *GithubSetting) accept(owner, name string) bool {
	// We regards the empty list as "Accept all incoming webhook".
	if g.acceptedRepos == nil {
		return true
	}

	k := owner + "/" + name
	_, ok := g.acceptedRepos[k]
	return ok
}
