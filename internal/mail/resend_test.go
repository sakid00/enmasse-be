package mail

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResendSendOTPUsesTemplate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer re_test" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("payload json: %v body=%s", err, body)
		}
		if got := payload["from"]; got != "MASSmaker <hello@enmasse.id>" {
			t.Errorf("from = %v", got)
		}
		to, _ := payload["to"].([]any)
		if len(to) != 1 || to[0] != "artist@example.com" {
			t.Errorf("to = %v", payload["to"])
		}
		if _, ok := payload["html"]; ok {
			t.Errorf("html should be omitted: %s", body)
		}
		tmpl, _ := payload["template"].(map[string]any)
		if tmpl["id"] != "tmpl_otp" {
			t.Errorf("template.id = %v", tmpl["id"])
		}
		vars, _ := tmpl["variables"].(map[string]any)
		if vars["code"] != "123456" {
			t.Errorf("code = %v", vars["code"])
		}
		if vars["TTLMinutes"] != float64(10) {
			t.Errorf("TTLMinutes = %v", vars["TTLMinutes"])
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"The enmasse.id domain is not verified."}`))
	}))
	t.Cleanup(srv.Close)

	m := NewResendMailer("re_test", "MASSmaker <hello@enmasse.id>", "tmpl_otp", 10*time.Minute)
	m.endpoint = srv.URL
	err := m.SendOTP(context.Background(), "artist@example.com", "123456")
	if err == nil {
		t.Fatal("expected resend error")
	}
	if !strings.Contains(err.Error(), "403") || !strings.Contains(err.Error(), "domain is not verified") {
		t.Fatalf("err = %v", err)
	}
}

func TestResendSendClaimUsesText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("payload json: %v body=%s", err, body)
		}
		if payload["subject"] != "Set your MASSmaker password" {
			t.Errorf("subject = %v", payload["subject"])
		}
		text, _ := payload["text"].(string)
		if !strings.Contains(text, "https://app.example/set-password?token=abc") {
			t.Errorf("text = %v", payload["text"])
		}
		if _, ok := payload["template"]; ok {
			t.Errorf("template should be omitted for claim: %s", body)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	t.Cleanup(srv.Close)

	m := NewResendMailer("re_test", "MASSmaker <hello@enmasse.id>", "tmpl_otp", 10*time.Minute)
	m.endpoint = srv.URL
	if err := m.SendClaim(context.Background(), "artist@example.com", "https://app.example/set-password?token=abc"); err != nil {
		t.Fatal(err)
	}
}
