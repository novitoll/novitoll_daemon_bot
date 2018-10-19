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

func (ingressBody *BotIngressRequest) Process(app *App) {
	log.Printf("[.] Process: Running. Message from -- Username: %s, Chat: %s, Message_Id: %d", ingressBody.Message.From.Username, ingressBody.Message.Chat.Username, ingressBody.Message.MessageId)

	job := &Job{ingressBody, app}

	results, errors := FanOutProcessJobs(job, []ProcessJobFn{
		JobNewChatMemberDetector,
		JobNewChatMemberWaiter,
		JobUrlDuplicationDetector,
		JobMessageStatistics,
	})

	completedProcessJobCount += 1

	for _, e := range errors {
		log.Fatalf("[-] %s", e.Error())
	}

	log.Printf("[+] %d. Process: Completed. Success/Failed=%d/%d", completedProcessJobCount, len(results), len(errors))
}