package bot

import (
	"log"
	"sync"
)

var (
	// total jobs count as long as the Go binary is running
	completedProcessJobCount = 0
)

type ProcessJobFn func(br *BotIngressRequest, rh *RouteHandler) (interface{}, error)

// br - BotIngressRequest (HTTP POST request body from Telegram)
// rh - RouteHander, struct which includes the bot's configuration (Telegram token) etc.
func FanOutProcessJobs(br *BotIngressRequest, rh *RouteHandler, jobs []ProcessJobFn) ([]interface{}, []error) {
	var wg sync.WaitGroup
	errJob := make(chan error, len(jobs))
	resultJob := make(chan interface{}, len(jobs))

	wg.Add(len(jobs))

	for _, v := range jobs {
		go func(job ProcessJobFn) {
			defer wg.Done()
			result, err := job(br, rh)
			if err != nil {
				errJob <- err
			} else {
				resultJob <- result
			}
		}(v)
	}

	wg.Wait()

	errJobs := make([]error, 0, len(jobs))
	resultJobs := make([]interface{}, 0, len(jobs))

	for range jobs {
		select {
		case r := <-resultJob:
			resultJobs = append(resultJobs, r)
		case e := <-errJob:
			errJobs = append(errJobs, e)
		}
	}
	return resultJobs, errJobs
}

func (br *BotIngressRequest) Process(rh *RouteHandler) {
	log.Printf("[.] Processing message from -- Username: %s, Chat: %s, Message_Id: %d", br.Message.From.Username, br.Message.Chat.Username, br.Message.MessageId)

	results, errors := FanOutProcessJobs(br, rh, []ProcessJobFn{
		JobNewcomerDetector,
		JobUrlDuplicationDetector,
	})

	completedProcessJobCount += 1

	log.Printf("[+] Process-%d: Completed. Success/Failed=%d/%d", completedProcessJobCount, len(results), len(errors))
}