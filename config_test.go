package main

import (
	"encoding/json"
	"testing"
)

func TestConfigUnmarshalNewPusherShape(t *testing.T) {
	payload := `{
		"listen": ":1111",
		"pushers": [
			{
				"name": "ops-wecom",
				"type": "wecom",
				"enabled": true,
				"config": {
					"corpid": "corp",
					"corpsecret": "secret",
					"agentid": 1000002,
					"touserdefault": "ops",
					"totagdefault": "grafana"
				}
			}
		]
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(payload), &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	cfg.normalize()

	if cfg.Listen != ":1111" {
		t.Fatalf("expected listen :1111, got %q", cfg.Listen)
	}
	if len(cfg.Pushers) != 1 {
		t.Fatalf("expected 1 pusher, got %d", len(cfg.Pushers))
	}
	pusher := cfg.Pushers[0]
	if pusher.Name != "ops-wecom" || pusher.Type != "wecom" || !pusher.IsEnabled() {
		t.Fatalf("unexpected pusher metadata: %+v", pusher)
	}
	wecomCfg, err := pusher.WeComConfig()
	if err != nil {
		t.Fatalf("wecom config: %v", err)
	}
	if wecomCfg.CorpID != "corp" || wecomCfg.AgentID != 1000002 || wecomCfg.ToTagDefault != "grafana" {
		t.Fatalf("unexpected wecom config: %+v", wecomCfg)
	}
}

func TestConfigNormalizeLegacyWeComShape(t *testing.T) {
	cfg := Config{
		Listen:        ":2222",
		CorpID:        "corp",
		CorpSecret:    "secret",
		AgentID:       1000003,
		ToUserDefault: "legacy-user",
	}

	cfg.normalize()

	if len(cfg.Pushers) != 1 {
		t.Fatalf("expected 1 migrated pusher, got %d", len(cfg.Pushers))
	}
	pusher := cfg.Pushers[0]
	if pusher.Type != "wecom" || !pusher.IsEnabled() {
		t.Fatalf("unexpected migrated pusher metadata: %+v", pusher)
	}
	wecomCfg, err := pusher.WeComConfig()
	if err != nil {
		t.Fatalf("wecom config: %v", err)
	}
	if wecomCfg.CorpID != "corp" || wecomCfg.CorpSecret != "secret" || wecomCfg.AgentID != 1000003 {
		t.Fatalf("unexpected migrated wecom config: %+v", wecomCfg)
	}
	if wecomCfg.ToUserDefault != "legacy-user" {
		t.Fatalf("expected legacy default user, got %q", wecomCfg.ToUserDefault)
	}
}

func TestNewPushServiceRejectsUnsupportedType(t *testing.T) {
	cfg := Config{
		Pushers: []PusherConfig{
			{
				Name: "unknown",
				Type: "unknown",
			},
		},
	}

	if _, err := NewPushService(cfg); err == nil {
		t.Fatal("expected unsupported pusher type error")
	}
}

func TestPushServiceSendWithoutEnabledPushers(t *testing.T) {
	disabled := false
	service, err := NewPushService(Config{
		Pushers: []PusherConfig{
			{
				Name:    "disabled-wecom",
				Type:    "wecom",
				Enabled: &disabled,
			},
		},
	})
	if err != nil {
		t.Fatalf("new push service: %v", err)
	}

	if err := service.Send("test", PushTarget{}); err == nil {
		t.Fatal("expected no enabled pushers error")
	}
}
