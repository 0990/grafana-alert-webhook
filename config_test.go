package main

import (
	"encoding/json"
	"testing"
)

func TestConfigUnmarshalTopLevelPushers(t *testing.T) {
	payload := `{
		"listen": ":1111",
		"wecom": {
			"enable": true,
			"corpid": "corp",
			"corpsecret": "secret",
			"agentid": 1000002,
			"touserdefault": "ops",
			"totagdefault": "grafana"
		},
		"xiaoshan": {
			"enable": true,
			"webhook_url": "https://example.com/robot/webhook/v2",
			"access_token": "token",
			"secret": "secret",
			"msg_type": "markdown",
			"at_all": true,
			"at_userids": ["u1"],
			"at_mobiles": ["13800000000"],
			"max_text_length": 1200
		}
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(payload), &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	cfg.normalize()

	if cfg.Listen != ":1111" {
		t.Fatalf("expected listen :1111, got %q", cfg.Listen)
	}
	if !cfg.WeCom.Enable {
		t.Fatal("expected wecom to be enabled")
	}
	if cfg.WeCom.CorpID != "corp" || cfg.WeCom.AgentID != 1000002 || cfg.WeCom.ToTagDefault != "grafana" {
		t.Fatalf("unexpected wecom config: %+v", cfg.WeCom)
	}
	if !cfg.XiaoShan.Enable {
		t.Fatal("expected xiaoshan to be enabled")
	}
	if cfg.XiaoShan.URL != "https://example.com/robot/webhook/v2" {
		t.Fatalf("unexpected xiaoshan url: %q", cfg.XiaoShan.URL)
	}
	if cfg.XiaoShan.AccessToken != "token" || cfg.XiaoShan.Secret != "secret" || cfg.XiaoShan.MsgType != "markdown" {
		t.Fatalf("unexpected xiaoshan config: %+v", cfg.XiaoShan)
	}
	if !cfg.XiaoShan.AtAll {
		t.Fatal("expected xiaoshan at all to be true")
	}
	if len(cfg.XiaoShan.AtUserIDs) != 1 || cfg.XiaoShan.AtUserIDs[0] != "u1" {
		t.Fatalf("unexpected xiaoshan at user ids: %+v", cfg.XiaoShan.AtUserIDs)
	}
	if len(cfg.XiaoShan.AtMobiles) != 1 || cfg.XiaoShan.AtMobiles[0] != "13800000000" {
		t.Fatalf("unexpected xiaoshan at mobiles: %+v", cfg.XiaoShan.AtMobiles)
	}
	if cfg.XiaoShan.MaxTextLength != 1200 {
		t.Fatalf("unexpected xiaoshan max text length: %d", cfg.XiaoShan.MaxTextLength)
	}
}

func TestConfigNormalizeDefaultListen(t *testing.T) {
	cfg := Config{}

	cfg.normalize()

	if cfg.Listen != ":1111" {
		t.Fatalf("expected default listen :1111, got %q", cfg.Listen)
	}
}

func TestNewPushServiceUsesTopLevelEnabledConfigs(t *testing.T) {
	service, err := NewPushService(Config{
		WeCom: WeComConfig{
			Enable: true,
		},
		XiaoShan: XiaoShanConfig{
			Enable: true,
		},
	})
	if err != nil {
		t.Fatalf("new push service: %v", err)
	}

	if len(service.pushers) != 2 {
		t.Fatalf("expected 2 pushers, got %d", len(service.pushers))
	}
	if service.pushers[0].typ != "wecom" || service.pushers[1].typ != "xiaoshan" {
		t.Fatalf("unexpected pushers: %+v", service.pushers)
	}
}

func TestPushServiceSendWithoutEnabledPushers(t *testing.T) {
	service, err := NewPushService(Config{})
	if err != nil {
		t.Fatalf("new push service: %v", err)
	}

	if err := service.Send("test", PushTarget{}); err == nil {
		t.Fatal("expected no enabled pushers error")
	}
}
