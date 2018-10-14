package bot

import (
	"log"
	"sync"
)

var (
	// total jobs count as long as the Go binary is running
	completedProcessJobCount = 0
)

type ProcessJobFn func(job *Job) (interface{}, error)

// br - BotIngressRequest (HTTP POST request body from Telegram)
// rh - RouteHander, struct which includes the bot's configuration (Telegram token) etc.

func FanOutProcessJobs(job *Job, jobsFn []ProcessJobFn) ([]interface{}, []error) {
	var wg sync.WaitGroup
	errJob := make(chan error, len(jobsFn))
	resultJob := make(chan interface{}, len(jobsFn))

	wg.Add(len(jobsFn))

	for _, v := range jobsFn {
		go func(jobFn ProcessJobFn) {
			defer wg.Done()
			result, err := jobFn(job) // here could not find the way to call job.jobFn(), so have to pass job struct as the parameter
			if err != nil {
				errJob <- err
			} else {
				resultJob <- result
			}
		}(v)
	}

	wg.Wait()

	errJobs := make([]error, 0, len(jobsFn))
	resultJobs := make([]interface{}, 0, len(jobsFn))

	for range jobsFn {
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

	job := &Job{br, rh}

	results, errors := FanOutProcessJobs(job, []ProcessJobFn{
		JobNewChatMemberDetector,
		JobUrlDuplicationDetector,
	})

	completedProcessJobCount += 1

	log.Printf("[+] %d. Process: Completed. Success/Failed=%d/%d", completedProcessJobCount, len(results), len(errors))
}