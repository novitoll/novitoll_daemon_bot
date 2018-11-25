// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
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
	app.Logger.WithFields(logrus.Fields{
		"userId":    ingressBody.Message.From.Id,
		"username":  ingressBody.Message.From.Username,
		"chat":      ingressBody.Message.Chat.Username,
		"messageId": ingressBody.Message.MessageId,
	}).Info("Process: Running.")

	job := &Job{ingressBody, app}

	results, errors := FanOutProcessJobs(job, []ProcessJobFn{
		JobNewChatMemberDetector,
		JobNewChatMemberWaiter,
		JobUrlDuplicationDetector,
		JobMessageStatistics,
		JobAdDetector,
		JobStickersDetector,
	})

	for _, e := range errors {
		app.Logger.Fatal(fmt.Sprintf("%v", e))
	}

	app.Logger.WithFields(logrus.Fields{
		"completed": len(results),
		"errors":    len(errors),
	}).Info("Process: Completed")
}
