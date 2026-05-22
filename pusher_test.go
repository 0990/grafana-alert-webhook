package main

import "testing"

type recordingPusher struct {
	messages []string
	targets  []PushTarget
}

func (p *recordingPusher) Send(text string, target PushTarget) error {
	p.messages = append(p.messages, text)
	p.targets = append(p.targets, target)
	return nil
}

func TestPushServiceSendByTypeOnlySendsMatchingPushers(t *testing.T) {
	wecom := &recordingPusher{}
	xiaoShan := &recordingPusher{}
	service := &PushService{
		pushers: []namedPusher{
			{typ: "wecom", name: "wecom", pusher: wecom},
			{typ: "xiaoshan", name: "xiaoshan", pusher: xiaoShan},
		},
	}

	if err := service.SendByType("xiaoshan", "hello", PushTarget{ToUser: "u1"}); err != nil {
		t.Fatalf("send by type: %v", err)
	}

	if len(wecom.messages) != 0 {
		t.Fatalf("expected wecom pusher not to be called, got %d calls", len(wecom.messages))
	}
	if len(xiaoShan.messages) != 1 || xiaoShan.messages[0] != "hello" {
		t.Fatalf("unexpected xiaoshan calls: %+v", xiaoShan.messages)
	}
	if xiaoShan.targets[0].ToUser != "u1" {
		t.Fatalf("unexpected target: %+v", xiaoShan.targets[0])
	}
}

func TestPushServiceSendByTypeRejectsMissingType(t *testing.T) {
	service := &PushService{
		pushers: []namedPusher{
			{typ: "wecom", name: "wecom", pusher: &recordingPusher{}},
		},
	}

	if err := service.SendByType("xiaoshan", "hello", PushTarget{}); err == nil {
		t.Fatal("expected missing pusher type error")
	}
}
