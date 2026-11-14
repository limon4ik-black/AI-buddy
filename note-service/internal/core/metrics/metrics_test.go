package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
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
