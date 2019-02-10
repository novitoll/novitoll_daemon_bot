// SPDX-License-Identifier: GPL-2.0
package config

type FeaturesCfg struct {
	UrlDuplication       UrlConfigOpt               `json:"urlDuplication"`
	NewcomerQuestionnare NewcomerConfigOpt          `json:"newcomerQuestionnare"`
	StickersDetection    StickersDetectionConfigOpt `json:"stickersDetection"`
	MsgStats             MsgStatsConfigOpt          `json:"messageStatistics"`
	Administration       AdministrationConfigOpt    `json:"administration"`
	AdDetection          AdDetectionConfigOpt       `json:"adDetection"`
}

type NewcomerConfigOpt struct {
	Enabled        bool  `json:"enabled"`
	ActionNotify   bool  `json:"actionNotify"`
	AuthTimeout    uint8 `json:"authTimeout"`
	KickBanTimeout uint8 `json:"kickBanTimeout"`
	I18n           map[string]struct {
		WelcomeMessage     string `json:"welcomeMessage"`
		AuthOKMessage      string `json:"authOKMessage"`
		AuthMessage        string `json:"authMessage"`
		AuthPasswd         string `json:"authPasswd"`
		AuthMessageCached  string `json:"authMessageCached"`
		AuthMessageURLPost string `json:"authMessageURLPost"`
	} `json:"i18n"`
}

type UrlConfigOpt struct {
	Enabled          bool `json:"enabled"`
	IgnoreHostnames  bool `json:"ignoreHostnames"`
	ActionNotify     bool `json:"actionNotify"`
	RelevanceTimeout int  `json:"relevanceTimeout"`
	I18n             map[string]struct {
		WarnMessage string `json:"warnMessage"`
	} `json:"i18n"`
}

type StickersDetectionConfigOpt struct {
	Enabled bool `json:"enabled"`
	I18n    map[string]struct {
		WarnMessage string `json:"warnMessage"`
	}
}

type MsgStatsConfigOpt struct {
	Enabled bool `json:"enabled"`
	I18n    map[string]struct {
		WarnMessageTooFreq  string `json:"warnMessageTooFreq"`
		WarnMessageTooLong  string `json:"warnMessageTooLong"`
		warnMessageDialogue string `json:"warnMessageDialogue"`
	} `json:"i18n"`
}

type AdministrationConfigOpt struct {
	LogLevel string `json:"logLevel"`
	I18n     map[string]struct {
		CronJobNewcomersReport string `json:"cronJobNewcomersReport"`
		CronJobUserMsgReport   string `json:"cronJobUserMsgReport"`
	} `json:"i18n"`
}

type AdDetectionConfigOpt struct {
	Enabled bool `json:"enabled"`
	I18n    map[string]struct {
		WarnMessage string `json:"warnMessage"`
	} `json:"i18n"`
}
