package bot

type Job struct {
	br *BotIngressRequest
	rh *RouteHandler
}

var NewComers map[int]interface{}
