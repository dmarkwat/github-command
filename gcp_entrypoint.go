package gcf

import (
	"github.com/dmarkwat/github-command/pkg/web"
	"log"
	"net/http"
	"os"
)

func HandleWebhook(resp http.ResponseWriter, req *http.Request) {
	webhookKey := os.Getenv("WEBHOOK_KEY")
	if webhookKey == "" {
		log.Printf("webhook key not provided!")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	web.HandleWebhook(resp, req, webhookKey)
}
