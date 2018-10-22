package bot

import (
	"regexp"
	"strings"
	"time"
)

// formula 1. (Incremental average) M_n = M_(n-1) + ((A_n - M_(n-1)) / n), where M_n = total mean, n = count of records, A = the array of elements

type UserMessageStats struct {
	MsgsLength 			[]int
	MeanMsgLength 		int
	// MeanMsgFrequency 	int
	MsgsCount 			int
	LastMsgTime 		int64
}

var (
	UserStatistics = make(map[int]*UserMessageStats)
	rgxPunct, _ = regexp.Compile("[^a-zA-Z0-9]+")
)

func JobMessageStatistics(job *Job) (interface{}, error) {
	if !job.HasMessageContent() {
		return nil, nil
	}

	// 1. get the prev. stats
	prev := UserStatistics[job.ingressBody.Message.From.Id]

	// 2. update the user stats
	wordsCount := getMessageLength(job.ingressBody.Message.Text)
	prev.MsgsLength = append(prev.MsgsLength, wordsCount)
	prev.LastMsgTime -= time.Now().Unix()
	prev.MsgsCount += 1
	prev.MeanMsgLength += ((wordsCount - prev.MeanMsgLength) / prev.MsgsCount)  // Ref:formula 1.
	
	// 3. update the user stats map
	UserStatistics[job.ingressBody.Message.From.Id] = prev

	return nil, nil
}

func getMessageLength(text string) int {
	// remove all non-alphanumerical chars and split per whitespace onto words
    words := strings.Fields(rgxPunct.ReplaceAllString(text, ""))
	return len(words)
}