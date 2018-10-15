package config

type FeaturesConfig struct {
	NotificationTarget	TargetOptions `json:"notificationTarget"`
	UrlDuplication		UrlConfigOptions `json:"urlDuplication"`
	NewcomerQuestionnare	NewcomerConfigOptions `json:"newcomerQuestionnare"`
}

type TargetOptions struct {
	Admins	[]string `json:"admins"`
}

type NewcomerConfigOptions struct {
    Enabled		bool `json:"enabled"`
    ActionKick	bool `json:"actionKick"`
    ActionBan	bool `json:"actionBan"`
    ActionNotify	bool `json:"actionNotify"`
    AuthMessage	string `json:"authMessage"`
    AuthTimeout	uint8 `json:"authTimeout"`
    KickBanTimeout uint8 `json:"kickBanTimeout"`
}

type UrlConfigOptions struct {
	Enabled		bool `json:"enabled"`
	IgnoreHostnames bool `json:"ignoreHostnames"`
	ActionNotify	bool `json:"actionNotify"`
}