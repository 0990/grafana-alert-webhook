package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type PushTarget struct {
	ToUser string
	ToTag  string
}

type Pusher interface {
	Send(text string, target PushTarget) error
}

type PushService struct {
	pushers []namedPusher
}

type namedPusher struct {
	name   string
	pusher Pusher
}

func NewPushService(cfg Config) (*PushService, error) {
	service := &PushService{}

	for _, pusherCfg := range cfg.Pushers {
		if !pusherCfg.IsEnabled() {
			continue
		}

		name := strings.TrimSpace(pusherCfg.Name)
		if name == "" {
			name = pusherCfg.Type
		}

		switch pusherCfg.Type {
		case "wecom":
			wecomCfg, err := pusherCfg.WeComConfig()
			if err != nil {
				return nil, fmt.Errorf("%s config: %w", name, err)
			}
			service.pushers = append(service.pushers, namedPusher{
				name:   name,
				pusher: &WXPusher{cfg: wecomCfg},
			})
		default:
			return nil, fmt.Errorf("unsupported pusher type %q", pusherCfg.Type)
		}
	}

	return service, nil
}

func (s *PushService) Send(text string, target PushTarget) error {
	if s == nil || len(s.pushers) == 0 {
		return errors.New("no enabled pushers configured")
	}

	var errs []error
	for _, pusher := range s.pushers {
		if err := pusher.pusher.Send(text, target); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", pusher.name, err))
		}
	}
	return errors.Join(errs...)
}

func (p PusherConfig) WeComConfig() (WeComConfig, error) {
	if len(p.Config) == 0 {
		return p.WeCom, nil
	}

	var cfg WeComConfig
	if err := json.Unmarshal(p.Config, &cfg); err != nil {
		return WeComConfig{}, err
	}
	return cfg, nil
}
