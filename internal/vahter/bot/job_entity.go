package bot

type Job struct {
	ingressBody *BotIngressRequest
	app *App
}

var (
	NewComers = make(map[int]interface{})
)