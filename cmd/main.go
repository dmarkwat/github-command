package main

import (
	"flag"
	"github.com/dmarkwat/github-command/pkg/web"
	"log"
	"net/http"
	"os"
)

type serverHandlers struct {
	webhookKey string
}

func (h serverHandlers) handleWebhook(resp http.ResponseWriter, req *http.Request) {
	web.HandleWebhook(resp, req, h.webhookKey)
}

func LookupEnvOrString(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	var webhookKey string

	flag.StringVar(&webhookKey, "webhook-key", LookupEnvOrString("WEBHOOK_KEY", ""), "webhook key github will use when sending requests")

	flag.Parse()

	if webhookKey == "" {
		log.Fatalf("no webhook key provided")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/command", serverHandlers{webhookKey: webhookKey}.handleWebhook)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
