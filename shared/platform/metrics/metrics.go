// Пакет metrics — метрики Prometheus для всех сервисов репы: свой реестр,
// счётчики, мидлварь для net/http и рекордер, которым пользуются сервисы
// на других роутерах (gin).
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry — свой реестр на процесс, а не глобальный: иначе тесты и повторные
// инициализации падают на «duplicate metrics collector registration».
var Registry = prometheus.NewRegistry()

var (
	requests = promauto.With(Registry).NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "HTTP-запросы по коду ответа",
	}, []string{"service", "method", "path", "status"})

	duration = promauto.With(Registry).NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Длительность HTTP-запросов",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "path"})

	inFlight = promauto.With(Registry).NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Запросы, которые обрабатываются прямо сейчас",
	}, []string{"service"})
)

// Handler отдаёт метрики в формате Prometheus.
func Handler() http.Handler {
	return promhttp.HandlerFor(Registry, promhttp.HandlerOpts{})
}

// InFlight отмечает начало запроса и возвращает функцию завершения.
func InFlight(service string) func() {
	inFlight.WithLabelValues(service).Inc()
	return func() { inFlight.WithLabelValues(service).Dec() }
}

// Observe фиксирует обработанный запрос. path — шаблон маршрута, а не сырой URL:
// иначе каждый id заводил бы новую серию метрик и Prometheus бы задохнулся.
// Пустой шаблон означает, что запрос не совпал ни с одним маршрутом.
func Observe(service, method, path, status string, seconds float64) {
	if path == "" {
		path = "unmatched"
	}
	duration.WithLabelValues(service, method, path).Observe(seconds)
	requests.WithLabelValues(service, method, path, status).Inc()
}

// Middleware считает запросы в сервисах на net/http. Шаблон маршрута берётся из
// r.Pattern, который проставляет ServeMux (Go 1.23+): конкретных id там уже нет.
// Перед вызовом шаблон сбрасывается, чтобы «не нашли маршрут» отличалось от «/».
func Middleware(service string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Свои скрейпы и опросы здоровья не считаем: иначе в графиках
			// всегда висит постоянная линия от них.
			if r.URL.Path == "/metrics" || r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}
			done := InFlight(service)
			start := time.Now()
			rec := &recorder{ResponseWriter: w, status: http.StatusOK}
			r.Pattern = ""
			next.ServeHTTP(rec, r)
			done()
			Observe(service, r.Method, r.Pattern, strconv.Itoa(rec.status), time.Since(start).Seconds())
		})
	}
}

// recorder запоминает код ответа для метрики.
type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Flush и Unwrap оставлены, чтобы через обёртку работали стриминг и Hijack —
// иначе через шлюз перестал бы идти потоковый ответ.
func (r *recorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
