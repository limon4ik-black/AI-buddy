package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/coder/websocket"
	dto "github.com/prometheus/client_model/go"
)

func metricValue(t *testing.T, name, service, status string) float64 {
	t.Helper()
	families, err := Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.Metric {
			labels := map[string]string{}
			for _, label := range metric.Label {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["service"] == service && (status == "" || labels["status"] == status) {
				if family.GetType() == dto.MetricType_GAUGE {
					return metric.GetGauge().GetValue()
				}
				return metric.GetCounter().GetValue()
			}
		}
	}
	t.Fatalf("metric %s for %s status %s not found", name, service, status)
	return 0
}

func TestWebSocketUpgradeSurvivesMetricsWrapper(t *testing.T) {
	server := httptest.NewServer(Middleware("websocket-test")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		defer conn.CloseNow()
		if err := conn.Write(r.Context(), websocket.MessageText, []byte("ok")); err != nil {
			t.Errorf("write: %v", err)
		}
	})))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, response, err := websocket.Dial(ctx, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("upgrade: %d", response.StatusCode)
	}
	_, body, err := conn.Read(ctx)
	if err != nil || string(body) != "ok" {
		t.Fatalf("websocket message %q: %v", body, err)
	}
}

func TestPanicReleasesInFlightAndPropagates(t *testing.T) {
	service := "panic-test"
	handler := Middleware(service)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("failure")
	}))
	func() {
		defer func() {
			if got := recover(); got != "failure" {
				t.Fatalf("panic changed or swallowed: %v", got)
			}
		}()
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/panic", nil))
	}()
	if got := metricValue(t, "http_requests_in_flight", service, ""); got != 0 {
		t.Fatalf("in-flight gauge leaked: %v", got)
	}
	if got := metricValue(t, "http_requests_total", service, "500"); got != 1 {
		t.Fatalf("panic not counted as failure: %v", got)
	}
}

func TestStatusMatchesCommittedResponse(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
		send func(http.ResponseWriter)
	}{
		{"explicit", 201, func(w http.ResponseWriter) { w.WriteHeader(201); w.WriteHeader(500) }},
		{"implicit", 200, func(w http.ResponseWriter) { _, _ = w.Write([]byte("ok")); w.WriteHeader(500) }},
		{"flush", 200, func(w http.ResponseWriter) { w.(http.Flusher).Flush(); w.WriteHeader(500) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := "status-" + tc.name
			handler := Middleware(service)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { tc.send(w) }))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", "/status", nil))
			if response.Code != tc.want {
				t.Fatalf("response: got %d want %d", response.Code, tc.want)
			}
			if tc.name == "implicit" && response.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
				t.Fatalf("wrapper prevented Content-Type detection: %v", response.Header())
			}
			if got := metricValue(t, "http_requests_total", service, strconv.Itoa(tc.want)); got != 1 {
				t.Fatalf("request count: %v", got)
			}
		})
	}
}
