package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGrafanaAlertSummaryOfficialPayload(t *testing.T) {
	payload := `{
		"receiver": "My Super Webhook",
		"status": "firing",
		"orgId": 1,
		"alerts": [
			{
				"status": "firing",
				"labels": {
					"alertname": "High memory usage",
					"team": "blue",
					"zone": "us-1"
				},
				"annotations": {
					"description": "The system has high memory usage",
					"summary": "This alert was triggered for zone us-1"
				},
				"startsAt": "2021-10-12T09:51:03.157076+02:00",
				"endsAt": "0001-01-01T00:00:00Z",
				"generatorURL": "https://play.grafana.org/alerting/1afz29v7z/edit",
				"fingerprint": "c6eadffa33fcdf37",
				"silenceURL": "https://play.grafana.org/alerting/silence/new",
				"dashboardURL": "",
				"panelURL": "",
				"values": {
					"B": 44.23943737541908,
					"C": 1
				}
			},
			{
				"status": "firing",
				"labels": {
					"alertname": "High CPU usage",
					"team": "blue",
					"zone": "eu-1"
				},
				"annotations": {
					"description": "The system has high CPU usage"
				},
				"generatorURL": "https://play.grafana.org/alerting/d1rdpdv7k/edit",
				"values": {
					"B": 44.23943737541908,
					"C": 1
				}
			}
		],
		"groupLabels": {},
		"commonLabels": {
			"team": "blue"
		},
		"commonAnnotations": {},
		"externalURL": "https://play.grafana.org/",
		"version": "1",
		"groupKey": "{}:{}",
		"truncatedAlerts": 0,
		"title": "[FIRING:2]  (blue)",
		"state": "alerting",
		"message": "**Firing**"
	}`

	var msg GrafanaAlertMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	summary, err := msg.Summary()
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	assertContains(t, summary, "[FIRING:2]  (blue)")
	assertContains(t, summary, "1. High memory usage (firing)")
	assertContains(t, summary, "Labels: team=blue zone=us-1")
	assertContains(t, summary, "Values: B=44.24 C=1")
	assertContains(t, summary, "Summary: This alert was triggered for zone us-1")
	assertContains(t, summary, "Description: The system has high memory usage")
	assertContains(t, summary, "URL: https://play.grafana.org/alerting/1afz29v7z/edit")
	assertContains(t, summary, "2. High CPU usage (firing)")
}

func TestGrafanaAlertSummaryFallbackTitle(t *testing.T) {
	msg := GrafanaAlertMsg{
		Status:       "firing",
		CommonLabels: map[string]string{"alertname": "Disk full"},
		Alerts: []GrafanaAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "Disk full",
					"instance":  "db-01",
				},
				Values: map[string]interface{}{
					"A": 93.0,
					"B": 93.456,
					"C": "critical",
				},
			},
		},
	}

	summary, err := msg.Summary()
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	assertContains(t, summary, "[FIRING:1] Disk full")
	assertContains(t, summary, "Alert: Disk full (firing)")
	assertContains(t, summary, "Labels: instance=db-01")
	assertContains(t, summary, "Values: A=93 B=93.46 C=critical")
}

func TestGrafanaAlertSummaryResolved(t *testing.T) {
	msg := GrafanaAlertMsg{
		Status: "resolved",
		Alerts: []GrafanaAlert{
			{
				Status:      "resolved",
				Labels:      map[string]string{"alertname": "CPU high"},
				Annotations: map[string]string{"summary": "CPU recovered"},
			},
		},
	}

	summary, err := msg.Summary()
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	assertContains(t, summary, "[RESOLVED:1] CPU high")
	assertContains(t, summary, "Alert: CPU high (resolved)")
	assertContains(t, summary, "Summary: CPU recovered")
}

func TestGrafanaAlertSummaryWithoutAlertsAllowsTitleMessage(t *testing.T) {
	msg := GrafanaAlertMsg{
		Title:   "Someone is testing the alert notification within Grafana.",
		Message: "Test notification",
	}

	summary, err := msg.Summary()
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	assertContains(t, summary, "Someone is testing the alert notification within Grafana.")
	assertContains(t, summary, "Test notification")
}

func TestLegacyPayloadIsInvalid(t *testing.T) {
	payload := `{
		"dashboardId": 1,
		"evalMatches": [{"value": 10, "metric": "cpu"}],
		"ruleName": "Legacy CPU",
		"ruleUrl": "http://grafana.example/rule",
		"state": "alerting"
	}`

	var msg GrafanaAlertMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		t.Fatalf("unmarshal legacy payload: %v", err)
	}

	if _, err := msg.Summary(); err == nil {
		t.Fatal("expected legacy payload to be invalid")
	}
}

func TestHandleGrafanaAlertRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/grafana_alert", strings.NewReader("{"))
	rec := httptest.NewRecorder()

	handleGrafanaAlert(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	assertContains(t, rec.Body.String(), "body unmarshal json error")
}

func TestHandleGrafanaAlertRejectsLegacyPayload(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/grafana_alert", strings.NewReader(`{
		"evalMatches": [{"value": 10, "metric": "cpu"}],
		"ruleName": "Legacy CPU",
		"state": "alerting"
	}`))
	rec := httptest.NewRecorder()

	handleGrafanaAlert(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	assertContains(t, rec.Body.String(), "invalid grafana webhook payload")
}

func assertContains(t *testing.T, value, expected string) {
	t.Helper()
	if !strings.Contains(value, expected) {
		t.Fatalf("expected summary to contain %q\nsummary:\n%s", expected, value)
	}
}
