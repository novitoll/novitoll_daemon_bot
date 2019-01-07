package bot

import (
	"sync"
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
