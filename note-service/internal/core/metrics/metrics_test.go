package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/middleware"
	"go.uber.org/zap"
)

// Проверяем главное свойство: в метрику попадает шаблон маршрута, а не конкретный id.
func TestMiddlewareKeepsRouteTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware("note-service"))
	r.GET("/api/v1/notes/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/metrics", gin.WrapH(Handler()))

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/notes/45b9e3e63bdc8f84eda36dfbb98096eb", nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := w.Body.String()

	if !strings.Contains(body, `path="/api/v1/notes/:id"`) {
		t.Fatalf("в метриках нет шаблона маршрута:\n%s", body)
	}
	if strings.Contains(body, "45b9e3e6") {
		t.Fatal("в метрики попал сырой id — это взорвёт число серий в Prometheus")
	}
	if !strings.Contains(body, `status="200"`) {
		t.Fatalf("в метриках нет кода ответа:\n%s", body)
	}
	if !strings.Contains(body, "http_request_duration_seconds_bucket") {
		t.Fatal("нет гистограммы длительности")
	}
}

// Recover must finish the error response before metrics records its status.
func TestRecoveredPanicCountsFailureAndReleasesInFlight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware("panic-note-test"), middleware.Recover(zap.NewNop()))
	r.GET("/panic", func(c *gin.Context) { panic("failure") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/panic", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("panic response status: %d", w.Code)
	}
	metrics := httptest.NewRecorder()
	Handler().ServeHTTP(metrics, httptest.NewRequest("GET", "/metrics", nil))
	body := metrics.Body.String()
	if !strings.Contains(body, `http_requests_total{method="GET",path="/panic",service="panic-note-test",status="500"} 1`) {
		t.Fatalf("recovered panic not counted as500:\n%s", body)
	}
	if !strings.Contains(body, `http_requests_in_flight{service="panic-note-test"} 0`) {
		t.Fatalf("in-flight leaked after recoveredpanic:\n%s", body)
	}
}

// Запрос без маршрута не должен оставлять пустой label.
func TestMiddlewareMarksUnmatched(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware("note-service"))
	r.GET("/metrics", gin.WrapH(Handler()))

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/нетакого/пути", nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if !strings.Contains(w.Body.String(), `path="unmatched"`) {
		t.Fatalf("непометил запрос без маршрута:\n%s", w.Body.String())
	}
}
