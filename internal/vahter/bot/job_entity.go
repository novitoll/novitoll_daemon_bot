package bot

type Job struct {
	br *BotIngressRequest
	rh *RouteHandler
}

var NewComers = make(map[int]interface{})