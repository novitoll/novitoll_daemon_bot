// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

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
		JobNewChatMemberAuth,
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
