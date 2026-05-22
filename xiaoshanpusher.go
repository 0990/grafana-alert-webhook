package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultXiaoShanURL = "https://wapi.zhimagame.net:6543/robot/webhook/v2"

type XiaoShanPusher struct {
	cfg XiaoShanConfig
}

func (p *XiaoShanPusher) Send(text string, target PushTarget) error {
	cfg := p.cfg.normalized()
	if cfg.AccessToken == "" {
		return errors.New("xiaoshan accesstoken is required")
	}
	if cfg.Secret == "" {
		return errors.New("xiaoshan secret is required")
	}

	atUserIDs := cfg.AtUserIDs
	if target.ToUser != "" {
		atUserIDs = splitList(target.ToUser)
	}
	if len(target.AtMobiles) > 0 {
		cfg.AtMobiles = target.AtMobiles
	}
	if target.AtAll {
		cfg.AtAll = true
	}

	switch cfg.MsgType {
	case "markdown":
		return p.sendMarkdown(cfg, text, atUserIDs)
	case "text":
		return p.sendText(cfg, text, atUserIDs)
	default:
		return fmt.Errorf("unsupported xiaoshan msgtype %q", cfg.MsgType)
	}
}

func (p *XiaoShanPusher) sendText(cfg XiaoShanConfig, text string, atUserIDs []string) error {
	parts := splitRunes(text, cfg.MaxTextLength)
	var errs []error
	for _, part := range parts {
		payload := xiaoShanTextPayload{
			MsgType: "text",
			Text: struct {
				Content string `json:"content"`
			}{Content: part},
			At: p.atPayload(cfg, atUserIDs),
		}
		if err := p.requestAPI(cfg, payload); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (p *XiaoShanPusher) sendMarkdown(cfg XiaoShanConfig, text string, atUserIDs []string) error {
	payload := xiaoShanMarkdownPayload{
		MsgType: "markdown",
		Markdown: struct {
			Text    string `json:"text"`
			PicURL  string `json:"picUrl"`
			PhotoID int    `json:"photoId"`
		}{Text: text},
		At: p.atPayload(cfg, atUserIDs),
	}
	return p.requestAPI(cfg, payload)
}

func (p *XiaoShanPusher) atPayload(cfg XiaoShanConfig, atUserIDs []string) xiaoShanAtPayload {
	return xiaoShanAtPayload{
		IsAtAll:   cfg.AtAll,
		AtUserIDs: atUserIDs,
		AtMobiles: cfg.AtMobiles,
	}
}

func (p *XiaoShanPusher) requestAPI(cfg XiaoShanConfig, payload interface{}) error {
	log.Printf("xiaoshan msg:%+v", payload)

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(p.signedURL(cfg), "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Msg     string `json:"msg"`
		ErrMsg  string `json:"errmsg"`
	}

	var ret result
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status:%d code:%d message:%s", resp.StatusCode, ret.Code, firstNonEmpty(ret.Message, ret.Msg, ret.ErrMsg))
	}
	if ret.Code != 200 {
		return fmt.Errorf("code:%d message:%s", ret.Code, firstNonEmpty(ret.Message, ret.Msg, ret.ErrMsg))
	}
	return nil
}

func (p *XiaoShanPusher) signedURL(cfg XiaoShanConfig) string {
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	sign := p.sign(cfg.Secret, timestamp)

	values := url.Values{}
	values.Set("access_token", cfg.AccessToken)
	values.Set("timestamp", timestamp)
	values.Set("sign", sign)
	return cfg.URL + "?" + values.Encode()
}

func (p *XiaoShanPusher) sign(secret, timestamp string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + secret))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (c XiaoShanConfig) normalized() XiaoShanConfig {
	if c.URL == "" {
		c.URL = defaultXiaoShanURL
	}
	if c.MsgType == "" {
		c.MsgType = "text"
	}
	if c.MaxTextLength <= 0 {
		c.MaxTextLength = 2000
	}
	c.MsgType = strings.ToLower(strings.TrimSpace(c.MsgType))
	c.AccessToken = strings.TrimSpace(c.AccessToken)
	c.Secret = strings.TrimSpace(c.Secret)
	c.AtUserIDs = compactStrings(c.AtUserIDs)
	c.AtMobiles = compactStrings(c.AtMobiles)
	return c
}

func (c *XiaoShanConfig) UnmarshalJSON(data []byte) error {
	type alias XiaoShanConfig
	var cfg alias
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	readStringAlias(raw, &cfg.URL, "endpoint", "webhookurl", "webhook_url")
	readStringAlias(raw, &cfg.AccessToken, "access_token", "accessToken", "token")
	readStringAlias(raw, &cfg.MsgType, "msg_type", "msgType")
	readBoolAlias(raw, &cfg.AtAll, "at_all", "atAll")
	readStringSliceAlias(raw, &cfg.AtMobiles, "at_mobiles", "atMobiles")
	readStringSliceAlias(raw, &cfg.AtUserIDs, "at_userids", "at_user_ids", "atUserIds", "atUserIDs")
	readIntAlias(raw, &cfg.MaxTextLength, "max_text_length", "maxTextLength")

	*c = XiaoShanConfig(cfg)
	return nil
}

type xiaoShanAtPayload struct {
	IsAtAll   bool     `json:"isAtAll"`
	AtUserIDs []string `json:"atUserIds"`
	AtMobiles []string `json:"atMobiles"`
}

type xiaoShanTextPayload struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At xiaoShanAtPayload `json:"at"`
}

type xiaoShanMarkdownPayload struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Text    string `json:"text"`
		PicURL  string `json:"picUrl"`
		PhotoID int    `json:"photoId"`
	} `json:"markdown"`
	At xiaoShanAtPayload `json:"at"`
}

func splitRunes(value string, limit int) []string {
	if value == "" {
		return []string{""}
	}
	if limit <= 0 {
		return []string{value}
	}

	runes := []rune(value)
	parts := make([]string, 0, len(runes)/limit+1)
	for len(runes) > limit {
		parts = append(parts, string(runes[:limit]))
		runes = runes[limit:]
	}
	parts = append(parts, string(runes))
	return parts
}

func splitList(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '|'
	})
	return compactStrings(fields)
}

func compactStrings(values []string) []string {
	result := values[:0]
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func readStringAlias(raw map[string]json.RawMessage, target *string, keys ...string) {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		var parsed string
		if json.Unmarshal(value, &parsed) == nil && strings.TrimSpace(parsed) != "" {
			*target = parsed
			return
		}
	}
}

func readStringSliceAlias(raw map[string]json.RawMessage, target *[]string, keys ...string) {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		var parsed []string
		if json.Unmarshal(value, &parsed) == nil && len(parsed) > 0 {
			*target = parsed
			return
		}
	}
}

func readBoolAlias(raw map[string]json.RawMessage, target *bool, keys ...string) {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		var parsed bool
		if json.Unmarshal(value, &parsed) == nil {
			*target = parsed
			return
		}
	}
}

func readIntAlias(raw map[string]json.RawMessage, target *int, keys ...string) {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		var parsed int
		if json.Unmarshal(value, &parsed) == nil && parsed > 0 {
			*target = parsed
			return
		}
	}
}
