package bot

type Job struct {
	ingressBody *BotIngressRequest
	app *App
}

func (job *Job) HasMessageContent() bool {
	// || job.ingressBody.Message.Sticker.FileId == ""
	return job.ingressBody.Message.Text != ""
}