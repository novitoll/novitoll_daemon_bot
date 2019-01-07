package bot

import (
	"fmt"
	"time"

	"github.com/novitoll/novitoll_daemon_bot/internal/utils"
	"github.com/sirupsen/logrus"
)

const (
	EVERY_LAST_SEC_7TH_DAY = 604799
)

var (
	PrevAuth int
	PrevKick int
)

func CronJobNewcomersCount(job *Job) (interface{}, error) {
	// blocks the goroutine for 1 week

	replyTextTpl := job.app.Features.Administration.I18n[job.app.Lang].CronJobNewcomersReport // for short reference

	select {
	case <-time.After(time.Duration(EVERY_LAST_SEC_7TH_DAY) * time.Second):

		authDiff := utils.CountDiffInPercent(PrevAuth, len(NewComersAuthVerified))
		kickDiff := utils.CountDiffInPercent(PrevKick, len(NewComersKicked))

		replyText := fmt.Sprintf(replyTextTpl, len(NewComersAuthVerified), authDiff, len(NewComersKicked), kickDiff)

		botEgressReq := &BotEgressSendMessage{
			ChatId:                job.ingressBody.Message.Chat.Id,
			Text:                  replyText,
			ParseMode:             ParseModeMarkdown,
			DisableWebPagePreview: true,
			DisableNotification:   true,
			ReplyToMessageId:      job.ingressBody.Message.MessageId,
			ReplyMarkup:           &BotForceReply{ForceReply: false, Selective: true},
		}
		// notify user about the flood limit
		resp, err := botEgressReq.EgressSendToTelegram(job.app)
		if err != nil {
			return nil, err
		}

		// reset maps
		NewComersAuthVerified = make(map[int]interface{})
		NewComersKicked = make(map[int]interface{})

		return resp, err
	}
}

func (app *App) StartCronJobs() {
	job := &Job{nil, app}

	results, errors := FanOutProcessJobs(job, []ProcessJobFn{
		CronJobNewcomersCount,
	})

	for _, e := range errors {
		app.Logger.Fatal(fmt.Sprintf("%v", e))
	}

	app.Logger.WithFields(logrus.Fields{
		"completed": len(results),
		"errors":    len(errors),
	}).Info("StartCronJobs: Completed")
}
