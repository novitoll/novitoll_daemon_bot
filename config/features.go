package config

type FeaturesConfig struct {
	NotificationTarget   TargetOptions                  `json:"notificationTarget"`
	UrlDuplication       UrlConfigOptions               `json:"urlDuplication"`
	NewcomerQuestionnare NewcomerConfigOptions          `json:"newcomerQuestionnare"`
	StickersDetection    StickersDetectionConfigOptions `json:"stickersDetection"`
}

type TargetOptions struct {
	Admins []string `json:"admins"`
}

type NewcomerConfigOptions struct {
	Enabled        bool  `json:"enabled"`
	ActionKick     bool  `json:"actionKick"`
	ActionBan      bool  `json:"actionBan"`
	ActionNotify   bool  `json:"actionNotify"`
	AuthTimeout    uint8 `json:"authTimeout"`
	KickBanTimeout uint8 `json:"kickBanTimeout"`
	I18n           map[string]struct {
		WelcomeMessage string `json:"welcomeMessage"`
		AuthOKMessage  string `json:"authOKMessage"`
		AuthMessage    string `json:"authMessage"`
	} `json:"i18n"`
}

type UrlConfigOptions struct {
	Enabled          bool `json:"enabled"`
	IgnoreHostnames  bool `json:"ignoreHostnames"`
	ActionNotify     bool `json:"actionNotify"`
	RelevanceTimeout int  `json:"relevanceTimeout"`
}

type StickersDetectionConfigOptions struct {
	Enabled      bool `json:"enabled"`
	ActionDelete bool `json:"actionDelete"`
	ActionNotify bool `json:"actionNotify"`
	I18n         map[string]struct {
		NotificationMessage string `json:"notificationMessage"`
	}
}
