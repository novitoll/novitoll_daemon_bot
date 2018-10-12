package bot

import (
	"log"
)

// func (j *Job) JobNewcomerDetector() (interface{}, error) {
func JobNewcomerDetector(br *BotIngressRequest, rh *RouteHandler) (interface{}, error) {
	// put the newcomer ID to the Redis for 48h expiration
	// before expiration notify admins that newcomers have not said a word
	if br.Message.NewComer.Username != "" {
		log.Printf("[!] NewComer detected")
	}
	return nil, nil
}
