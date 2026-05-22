package main

import (
	"encoding/json"
	"flag"
	"io"
	log "log"
	"net/http"
	"strconv"

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

	http.HandleFunc("/wecom/grafana_alert", handleGrafanaAlertByType("wecom"))
	http.HandleFunc("/wecom/send", handleSendByType("wecom"))
	http.HandleFunc("/xiaoshan/grafana_alert", handleGrafanaAlertByType("xiaoshan"))
	http.HandleFunc("/xiaoshan/send", handleSendByType("xiaoshan"))

	err = http.ListenAndServe(cfg.Listen, nil)
	log.Print(err)
}

func handleGrafanaAlertByType(pusherType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleGrafanaAlertPush(pusherType, w, r)
	}
}

func handleGrafanaAlertPush(pusherType string, w http.ResponseWriter, r *http.Request) {
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

	target := requestPushTarget(r)

	err = pushService.SendByType(pusherType, message, target)
	if err != nil {
		log.Print(err)
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}

	respond(w, 0, "ok")
}

func handleSendByType(pusherType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleSendPush(pusherType, w, r)
	}
}

func handleSendPush(pusherType string, w http.ResponseWriter, r *http.Request) {
	content := r.FormValue("content")
	target := requestPushTarget(r)
	err := pushService.SendByType(pusherType, content, target)
	if err != nil {
		log.Print(err)
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}
	respond(w, 0, "ok")
}

func requestPushTarget(r *http.Request) PushTarget {
	atAll, _ := strconv.ParseBool(r.FormValue("atall"))
	return PushTarget{
		ToUser:    r.FormValue("touser"),
		ToTag:     r.FormValue("totag"),
		AtAll:     atAll,
		AtMobiles: splitList(r.FormValue("atmobiles")),
	}
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
