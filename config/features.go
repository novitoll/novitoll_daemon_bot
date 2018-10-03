package config

type FeaturesConfig struct {
	urlDuplication 			configOptions
	newcomerQuestionnare	configOptions
}

type configOptions struct {
    enabled 			bool
    actionKick 			bool
    actionBan 			bool
    actionAdminNotify 	bool
}
