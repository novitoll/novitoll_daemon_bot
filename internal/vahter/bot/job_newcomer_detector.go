package bot

import (
	"log"
)

func JobNewcomerDetector(j *Job) (interface{}, error) {
	// put the newcomer ID to the Redis for 48h expiration
	// before expiration notify admins that newcomers have not said a word
	if j.br.Message.NewComer.Username != "" {
		log.Printf("[!] NewComer detected")
	}
	return nil, nil
}
