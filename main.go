package main

import (
	"encoding/json"
	"flag"
	"io"
	log "log"
	"net/http"

	"github.com/sirupsen/logrus"
)

var confFile = flag.String("c", "config.json", "config file")

var pushService *PushService

func main() {
	flag.Parse()

	cfg, err := readOrCreateCfg(*confFile)
	if err != nil {
		log.Print(err)
		return
	}

	pushService, err = NewPushService(*cfg)
	if err != nil {
		log.Print(err)
		return
	}

	http.HandleFunc("/grafana_alert", handleGrafanaAlert)
	http.HandleFunc("/send", handleSend)

	err = http.ListenAndServe(cfg.Listen, nil)
	log.Print(err)
}

func handleGrafanaAlert(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "body read error")
		return
	}

	var alertMsg GrafanaAlertMsg
	err = json.Unmarshal(data, &alertMsg)
	if err != nil {
		respondError(w, http.StatusBadRequest, "body unmarshal json error")
		return
	}

	message, err := alertMsg.Summary()
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	toUser := r.FormValue("touser")
	toTag := r.FormValue("totag")
	target := PushTarget{ToUser: toUser, ToTag: toTag}

	err = pushService.Send(message, target)
	if err != nil {
		log.Print(err)
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}

	respond(w, 0, "ok")
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	content := r.FormValue("content")
	toUser := r.FormValue("touser")
	toTag := r.FormValue("totag")
	target := PushTarget{ToUser: toUser, ToTag: toTag}
	err := pushService.Send(content, target)
	if err != nil {
		log.Print(err)
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}
	respond(w, 0, "ok")
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	respond(w, -1, message)
}

func respond(w http.ResponseWriter, code int, message string) {
	type result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	data := result{
		Code:    code,
		Message: message,
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logrus.WithField("data", data).WithError(err).Error("Server respond json encode error")
	}
}
