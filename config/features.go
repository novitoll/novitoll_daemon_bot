package config

type FeaturesConfig struct {
	NotificationTarget	TargetOptions `json:"notificationTarget"`
	UrlDuplication		ConfigOptions `json:"urlDuplication"`
	NewcomerQuestionnare	ConfigOptions `json:"newcomerQuestionnare"`
	AdDetection		ConfigOptions `json:"adDetection"`
}

type TargetOptions struct {
	Admins	[]string `json:"admins"`
}

type ConfigOptions struct {
    Enabled		bool `json:"enabled"`
    ActionKick		bool `json:"actionKick"`
    ActionBan		bool `json:"actionBan"`
    ActionAdminNotify	bool `json:"actionAdminNotify"`
}
