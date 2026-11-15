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
			defer func() {
				done()
				if failure := recover(); failure != nil {
					// net/http closes the connection on an unhandled panic; count the
					// failure without swallowing it or sending a second response.
					Observe(service, r.Method, r.Pattern, "500", time.Since(start).Seconds())
					panic(failure)
				}
				Observe(service, r.Method, r.Pattern, strconv.Itoa(rec.status), time.Since(start).Seconds())
			}()
			r.Pattern = ""
			next.ServeHTTP(rec, r)
		})
	}
}

// recorder запоминает код ответа для метрики.
type recorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *recorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	// Informational responses can precede the final status. 101 is final:
	// WebSocket upgrades must retain their status in the metrics.
	if code >= 100 && code < 200 && code != http.StatusSwitchingProtocols {
		r.ResponseWriter.WriteHeader(code)
		return
	}
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(body []byte) (int, error) {
	if !r.wroteHeader {
		r.status = http.StatusOK
		r.wroteHeader = true
	}
	return r.ResponseWriter.Write(body)
}

// Flush и Unwrap оставлены, чтобы через обёртку работали стриминг и Hijack —
// иначе через шлюз перестал бы идти потоковый ответ.
func (r *recorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		if !r.wroteHeader {
			r.status = http.StatusOK
			r.wroteHeader = true
		}
		f.Flush()
	}
}

func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
