package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type WXPusher struct{
	cfg Config
}

func(p *WXPusher)Send(text string,toUser string,toTag string)error{
	if toUser==""&&toTag==""{
		toUser = p.cfg.ToUserDefault
	}

	token,err:=p.getAccessToken()
	if err!=nil{
		return err
	}

	return p.send(token,text,toUser,toTag)
}

func (p *WXPusher)getAccessToken()(string,error){
	url:=fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",p.cfg.CorpID,p.cfg.CorpSecret)
	resp,err:=http.Get(url)
	if err!=nil{
		return "",err
	}

	defer resp.Body.Close()

	type Result struct{
		ErrCode int `json:"errcode"`
		ErrMsg string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn int `json:"expires_in"`
	}

	var ret Result
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err!=nil{
		return "",err
	}

	if ret.ErrCode!=0{
		return "",fmt.Errorf("errcode:%d errmsg:%s",ret.ErrCode,ret.ErrMsg)
	}

	return ret.AccessToken,nil
}

func (p *WXPusher)send(accessToken,text,toUser,toTag string)error{
	log.Println(fmt.Sprintf("msg:%s toUser:%s toTag:%s",text,toUser,toTag))

	type Request struct {
		Touser  string `json:"touser"`
		Toparty string `json:"toparty"`
		Totag   string `json:"totag"`
		Msgtype string `json:"msgtype"`
		Agentid int    `json:"agentid"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
		Safe                   int `json:"safe"`
		EnableIdTrans          int `json:"enable_id_trans"`
		EnableDuplicateCheck   int `json:"enable_duplicate_check"`
		DuplicateCheckInterval int `json:"duplicate_check_interval"`
	}

	req:=&Request{
		Touser:  toUser,
		Totag:   toTag,
		Msgtype: "text",
		Agentid: p.cfg.AgentId,
		Text: struct {
			Content string `json:"content"`
		}{
			Content: text,
		},
	}

	data,err:=json.Marshal(req)
	if err!=nil{
		return err
	}

	url:=fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s",accessToken)

	resp,err:=http.Post(url,"application/json",bytes.NewReader(data))
	if err!=nil{
		return err
	}

	defer resp.Body.Close()

	type Result struct {
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
		Invaliduser  string `json:"invaliduser"`
		Invalidparty string `json:"invalidparty"`
		Invalidtag   string `json:"invalidtag"`
	}

	var ret Result
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err!=nil{
		return err
	}

	if ret.ErrCode!=0{
		return fmt.Errorf("errcode:%d errmsg:%s",ret.ErrCode,ret.ErrMsg)
	}

	return nil
}
