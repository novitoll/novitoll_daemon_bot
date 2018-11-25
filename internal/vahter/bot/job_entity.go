// SPDX-License-Identifier: GPL-2.0
package bot

type Job struct {
	ingressBody *BotIngressRequest
	app         *App
}

func (job *Job) HasMessageContent() bool {
	return job.ingressBody.Message.Text != "" || job.ingressBody.Message.Sticker.FileId != ""
}

const (
	TIME_TO_DELETE_REPLY_MSG = 10
)
