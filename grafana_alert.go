package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type GrafanaAlertMsg struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	OrgID             int               `json:"orgId"`
	Alerts            []GrafanaAlert    `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
	Title             string            `json:"title"`
	State             string            `json:"state"`
	Message           string            `json:"message"`
}

type GrafanaAlert struct {
	Status       string                 `json:"status"`
	Labels       map[string]string      `json:"labels"`
	Annotations  map[string]string      `json:"annotations"`
	StartsAt     string                 `json:"startsAt"`
	EndsAt       string                 `json:"endsAt"`
	GeneratorURL string                 `json:"generatorURL"`
	Fingerprint  string                 `json:"fingerprint"`
	SilenceURL   string                 `json:"silenceURL"`
	DashboardURL string                 `json:"dashboardURL"`
	PanelURL     string                 `json:"panelURL"`
	ImageURL     string                 `json:"imageURL"`
	Values       map[string]interface{} `json:"values"`
}

func (p *GrafanaAlertMsg) Summary() (string, error) {
	if len(p.Alerts) == 0 && strings.TrimSpace(p.Title) == "" && strings.TrimSpace(p.Message) == "" {
		return "", errors.New("invalid grafana webhook payload: missing alerts, title, and message")
	}

	lines := []string{p.summaryTitle()}
	if len(p.Alerts) == 0 {
		if message := strings.TrimSpace(p.Message); message != "" {
			lines = append(lines, message)
		}
		return strings.Join(lines, "\n"), nil
	}

	for i, alert := range p.Alerts {
		if len(p.Alerts) > 1 {
			lines = append(lines, p.alertSummaryLine(i, alert))
		}
		if summary := annotationText(alert.Annotations); summary != "" {
			lines = append(lines, summary)
		}
		if values := formatValues(alert.Values); values != "" {
			lines = append(lines, "Values: "+values)
		}
		if labels := formatLabels(alert.Labels); labels != "" {
			lines = append(lines, "Labels:\n"+labels)
		}
	}

	if p.TruncatedAlerts > 0 {
		lines = append(lines, fmt.Sprintf("Truncated alerts: %d", p.TruncatedAlerts))
	}

	return strings.Join(lines, "\n"), nil
}

func (p *GrafanaAlertMsg) summaryTitle() string {
	if title := strings.TrimSpace(p.Title); title != "" && len(p.Alerts) == 0 {
		return title
	}

	status := strings.ToUpper(strings.TrimSpace(p.Status))
	if status == "" {
		status = strings.ToUpper(strings.TrimSpace(p.State))
	}
	if status == "" {
		status = "UNKNOWN"
	}

	name := firstNonEmpty(
		mapValue(p.CommonLabels, "alertname"),
		firstAlertName(p.Alerts),
		strings.TrimSpace(p.Receiver),
		"Grafana Alert",
	)

	if len(p.Alerts) > 0 {
		return fmt.Sprintf("[%s:%d] %s", status, len(p.Alerts), name)
	}
	return fmt.Sprintf("[%s] %s", status, name)
}

func (p *GrafanaAlertMsg) alertSummaryLine(index int, alert GrafanaAlert) string {
	name := firstNonEmpty(
		mapValue(alert.Labels, "alertname"),
		mapValue(p.CommonLabels, "alertname"),
		fmt.Sprintf("Alert %d", index+1),
	)
	status := firstNonEmpty(strings.TrimSpace(alert.Status), strings.TrimSpace(p.Status), "unknown")

	if len(p.Alerts) == 1 {
		return fmt.Sprintf("Alert: %s (%s)", name, status)
	}
	return fmt.Sprintf("%d. %s (%s)", index+1, name, status)
}

func firstAlertName(alerts []GrafanaAlert) string {
	for _, alert := range alerts {
		if name := mapValue(alert.Labels, "alertname"); name != "" {
			return name
		}
	}
	return ""
}

func annotationText(annotations map[string]string) string {
	summary := strings.TrimSpace(annotations["summary"])
	description := strings.TrimSpace(annotations["description"])

	switch {
	case summary != "" && description != "" && summary != description:
		return "Summary: " + summary + "\nDescription: " + description
	case summary != "":
		return "Summary: " + summary
	case description != "":
		return "Description: " + description
	default:
		return ""
	}
}

func formatLabels(values map[string]string) string {
	hidden := map[string]struct{}{
		"alertname":      {},
		"grafana_folder": {},
		"job":            {},
		"mountpoint":     {},
	}

	if len(values) == 0 {
		return ""
	}

	keys := make([]string, 0, len(values))
	for key, value := range values {
		if _, ok := hidden[key]; ok || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values[key])
	}
	return strings.Join(parts, "\n")
}

func formatValues(values map[string]interface{}) string {
	if len(values) == 0 {
		return ""
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+formatValue(values[key]))
	}
	return strings.Join(parts, " ")
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case float64:
		return formatFloat(v)
	case float32:
		return formatFloat(float64(v))
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case json.Number:
		return v.String()
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatFloat(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Sprintf("%v", value)
	}
	if math.Trunc(value) == value {
		return strconv.FormatFloat(value, 'f', 0, 64)
	}
	formatted := strconv.FormatFloat(value, 'f', 2, 64)
	formatted = strings.TrimRight(formatted, "0")
	return strings.TrimRight(formatted, ".")
}

func mapValue(values map[string]string, key string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[key])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
