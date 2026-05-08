package main

import (
	"encoding/json"
	"flag"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	log "log"
	"net/http"
)

var confFile = flag.String("c", "config.json", "config file")

var wxPusher *WXPusher

func main(){
	flag.Parse()

	cfg,err:=readOrCreateCfg(*confFile)
	if err!=nil{
		log.Print(err)
		return
	}

	wxPusher = &WXPusher{cfg: *cfg}

	http.HandleFunc("/grafana_alert",handleGrafanaAlert)
	http.HandleFunc("/send",handleSend)

	err=http.ListenAndServe(cfg.Listen,nil)
	log.Print(err)
}

func handleGrafanaAlert(w http.ResponseWriter, r *http.Request){
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respond(w, -1, "body read error")
		return
	}

	var alertMsg GrafanaAlertMsg
	err = json.Unmarshal(data, &alertMsg)
	if err != nil {
		respond(w, -1, "body unmarshal json error")
		return
	}

	title := alertMsg.SimpleTitle()
	//detail := alertMsg.Detail()

	toUser:=r.FormValue("touser")
	toTag:=r.FormValue("totag")

	err = wxPusher.Send(title,toUser,toTag)
	if err!=nil{
		log.Print(err)

	}


}

func handleSend(w http.ResponseWriter, r *http.Request){
	content:=r.FormValue("content")
	toUser:=r.FormValue("touser")
	toTag:=r.FormValue("totag")
	err := wxPusher.Send(content,toUser,toTag)
	if err!=nil{
		log.Print(err)
	}
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
