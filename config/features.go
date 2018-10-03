package config

type FeaturesConfig struct {
	UrlDuplication 			ConfigOptions `json:"urlDuplication"`
	NewcomerQuestionnare	ConfigOptions `json:"newcomerQuestionnare"`
}

type ConfigOptions struct {
    Enabled 			bool `json:"enabled"`
    ActionKick 			bool `json:"actionKick"`
    ActionBan 			bool `json:"actionBan"`
    ActionAdminNotify 	bool `json:"actionAdminNotify"`
}