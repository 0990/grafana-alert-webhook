package main

import (
	"errors"
	"fmt"
	"strings"
)

type PushTarget struct {
	ToUser    string
	ToTag     string
	AtAll     bool
	AtMobiles []string
}

type Pusher interface {
	Send(text string, target PushTarget) error
}

type PushService struct {
	pushers []namedPusher
}

type namedPusher struct {
	typ    string
	name   string
	pusher Pusher
}

func NewPushService(cfg Config) (*PushService, error) {
	service := &PushService{}

	if cfg.WeCom.Enable {
		service.pushers = append(service.pushers, namedPusher{
			typ:    "wecom",
			name:   "wecom",
			pusher: &WXPusher{cfg: cfg.WeCom},
		})
	}
	if cfg.XiaoShan.Enable {
		service.pushers = append(service.pushers, namedPusher{
			typ:    "xiaoshan",
			name:   "xiaoshan",
			pusher: &XiaoShanPusher{cfg: cfg.XiaoShan},
		})
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

func (s *PushService) SendByType(pusherType, text string, target PushTarget) error {
	if strings.TrimSpace(pusherType) == "" {
		return s.Send(text, target)
	}
	if s == nil || len(s.pushers) == 0 {
		return errors.New("no enabled pushers configured")
	}

	pusherType = strings.ToLower(strings.TrimSpace(pusherType))
	var matched int
	var errs []error
	for _, pusher := range s.pushers {
		if pusher.typ != pusherType {
			continue
		}
		matched++
		if err := pusher.pusher.Send(text, target); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", pusher.name, err))
		}
	}
	if matched == 0 {
		return fmt.Errorf("no enabled pushers configured for type %q", pusherType)
	}
	return errors.Join(errs...)
}
