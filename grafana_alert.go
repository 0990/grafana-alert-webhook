package main

import "fmt"

type GrafanaAlertMsg struct {
	DashboardID int `json:"dashboardId"`
	EvalMatches []struct {
		Value  int         `json:"value"`
		Metric string      `json:"metric"`
		Tags   interface{} `json:"tags"`
	} `json:"evalMatches"`
	Message  string `json:"message"`
	OrgID    int    `json:"orgId"`
	PanelID  int    `json:"panelId"`
	RuleID   int    `json:"ruleId"`
	RuleName string `json:"ruleName"`
	RuleURL  string `json:"ruleUrl"`
	State    string `json:"state"`
	Tags     struct {
	} `json:"tags"`
	Title string `json:"title"`
}

func (p *GrafanaAlertMsg) SimpleTitle() string {
	var title string
	title += "[" + p.State + "]"
	title += p.RuleName
	title += ","

	for _, v := range p.EvalMatches {
		//title += fmt.Sprintf("数量:%d", v.Value)
		title += fmt.Sprintf("%s:%d ", v.Metric, v.Value)
	}
	return title
}

func (p *GrafanaAlertMsg) Detail() string {
	detail := p.Message + " "
	for _, v := range p.EvalMatches {
		detail += fmt.Sprintf("%s:%d", v.Metric, v.Value)
	}
	return detail
}
