package prometheusutils

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var defaultMetricPath = "/metrics"
var defaultSubsystem = ""

const (
	_          = iota // ignore first value by assigning to blank identifier
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
)

// reqDurBuckets is the buckets for request duration. Here, we use the prometheus defaults
// which are for ~10s request length max: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
var ReqDurBuckets = []float64{0.001, 0.001048576, 0.001398101, 0.001747626, 0.002097151, 0.002446676,
	0.002796201, 0.003145726, 0.003495251, 0.003844776, 0.004194304, 0.005592405, 0.006990506,
	0.008388607, 0.009786708, 0.011184809, 0.01258291, 0.013981011, 0.015379112, 0.016777216,
	0.022369621, 0.027962026, 0.033554431, 0.039146836, 0.044739241, 0.050331646, 0.055924051,
	0.061516456, 0.067108864, 0.089478485, 0.111848106, 0.134217727, 0.156587348, 0.178956969,
	0.20132659, 0.223696211, 0.246065832, 0.268435456, 0.357913941, 0.447392426, 0.536870911,
	0.626349396, 0.715827881, 0.805306366, 0.894784851, 0.984263336, 1.073741824, 1.431655765,
	1.789569706, 2.147483647, 2.505397588, 2.863311529, 3.22122547, 3.579139411, 3.937053352,
	4.294967296, 5.726623061, 7.158278826, 8.589934591, 10.021590356, 11.453246121, 12.884901886,
	14.316557651, 15.748213416, 17.179869184, 22.906492245, 28.633115306, 30}

// reqSzBuckets is the buckets for request size. Here we define a spectrom from 1KB thru 1NB up to 10MB.
var ReqSzBuckets = []float64{1.0 * KB, 2.0 * KB, 5.0 * KB, 10.0 * KB, 100 * KB, 500 * KB, 1.0 * MB, 2.5 * MB, 5.0 * MB, 10.0 * MB}

// resSzBuckets is the buckets for response size. Here we define a spectrom from 1KB thru 1NB up to 10MB.
var ResSzBuckets = []float64{1.0 * KB, 2.0 * KB, 5.0 * KB, 10.0 * KB, 100 * KB, 500 * KB, 1.0 * MB, 2.5 * MB, 5.0 * MB, 10.0 * MB}

// Standard default metrics
//	counter, counter_vec, gauge, gauge_vec,
//	histogram, histogram_vec, summary, summary_vec
var reqCnt = &Metric{
	ID:          "reqCnt",
	Name:        "requests_total",
	Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	Type:        "counter_vec",
	Args:        []string{"code", "method", "host", "url"},
}

var reqDur = &Metric{
	ID:          "reqDur",
	Name:        "request_duration_seconds",
	Description: "The HTTP request latencies in seconds.",
	Args:        []string{"code", "method", "url"},
	Type:        "histogram_vec",
	Buckets:     ReqDurBuckets,
}

var resSz = &Metric{
	ID:          "resSz",
	Name:        "response_size_bytes",
	Description: "The HTTP response sizes in bytes.",
	Args:        []string{"code", "method", "url"},
	Type:        "histogram_vec",
	Buckets:     ResSzBuckets,
}

var reqSz = &Metric{
	ID:          "reqSz",
	Name:        "request_size_bytes",
	Description: "The HTTP request sizes in bytes.",
	Args:        []string{"code", "method", "url"},
	Type:        "histogram_vec",
	Buckets:     ReqSzBuckets,
}

var standardMetrics = []*Metric{
	reqCnt,
	reqDur,
	resSz,
	reqSz,
}

/*
RequestCounterLabelMappingFunc is a function which can be supplied to the middleware to control
the cardinality of the request counter's "url" label, which might be required in some contexts.
For instance, if for a "/customer/:name" route you don't want to generate a time series for every
possible customer name, you could use this function:
func(c echo.Context) string {
	url := c.Request.URL.Path
	for _, p := range c.Params {
		if p.Key == "name" {
			url = strings.Replace(url, p.Value, ":name", 1)
			break
		}
	}
	return url
}
which would map "/customer/alice" and "/customer/bob" to their template "/customer/:name".
It can also be applied for the "Host" label
*/
type RequestCounterLabelMappingFunc func(c *gin.Context) string

/*
Skipper defines a function to skip middleware. Returning true skips processing
the middleware.
*/
type Skipper func(*gin.Context) bool

func DefaultSkipper(*gin.Context) bool {
	return false
}

/*
Metric is a definition for the name, description, type, ID, and
prometheus.Collector type (i.e. CounterVec, Summary, etc) of each metric
*/
type Metric struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            string
	Args            []string
	Buckets         []float64
}

/*
Prometheus contains the metrics gathered by the instance and its path
*/
type Prometheus struct {
	reqCnt               *prometheus.CounterVec
	reqDur, reqSz, resSz *prometheus.HistogramVec

	MetricsList []*Metric
	MetricsPath string
	Subsystem   string
	Skipper     Skipper

	RequestCounterURLLabelMappingFunc  RequestCounterLabelMappingFunc
	RequestCounterHostLabelMappingFunc RequestCounterLabelMappingFunc

	// Context string to use as a prometheus URL label
	URLLabelFromContext string
}

/*
NewPrometheus generates a new set of metrics with a certain subsystem name
*/
func NewPrometheus(subsystem string, skipper Skipper, customMetricsList ...[]*Metric) *Prometheus {
	var metricsList []*Metric
	if skipper == nil {
		skipper = DefaultSkipper
	}

	if len(customMetricsList) > 1 {
		panic("Too many args. NewPrometheus( string, <optional []*Metric> ).")
	} else if len(customMetricsList) == 1 {
		metricsList = customMetricsList[0]
	}

	metricsList = append(metricsList, standardMetrics...)

	p := &Prometheus{
		MetricsList: metricsList,
		MetricsPath: defaultMetricPath,
		Subsystem:   defaultSubsystem,
		Skipper:     skipper,
		RequestCounterURLLabelMappingFunc: func(c *gin.Context) string {
			return c.Request.URL.Path // i.e. by default do nothing, i.e. return URL as is
		},
		RequestCounterHostLabelMappingFunc: func(c *gin.Context) string {
			return c.Request.Host
		},
	}

	p.registerMetrics(subsystem)

	return p
}

/*
NewMetric associates prometheus.Collector based on Metric.Type
*/
func NewMetric(m *Metric, subsystem string) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "counter":
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "gauge_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "gauge":
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
				Buckets:   m.Buckets,
			},
			m.Args,
		)
	case "histogram":
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
				Buckets:   m.Buckets,
			},
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "summary":
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}
	return metric
}

func (p *Prometheus) registerMetrics(subsystem string) {

	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, subsystem)
		if err := prometheus.Register(metric); err != nil {
			log.Printf("%s could not be registered in Prometheus: %v", metricDef.Name, err)
		}
		switch metricDef {
		case reqCnt:
			p.reqCnt = metric.(*prometheus.CounterVec)
		case reqDur:
			p.reqDur = metric.(*prometheus.HistogramVec)
		case resSz:
			p.resSz = metric.(*prometheus.HistogramVec)
		case reqSz:
			p.reqSz = metric.(*prometheus.HistogramVec)
		}
		metricDef.MetricCollector = metric
	}
}

/*
Use adds the middleware to the Echo engine.
*/
func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.HandlerFunc())
}

/*
HandlerFunc defines handler function for middleware
*/
func (p *Prometheus) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == p.MetricsPath {
			c.Next()
			return
		}
		if p.Skipper(c) {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		status := c.Writer.Status()

		elapsed := float64(time.Since(start)) / float64(time.Second)

		url := p.RequestCounterURLLabelMappingFunc(c)
		if len(p.URLLabelFromContext) > 0 {
			u, _ := c.Get(p.URLLabelFromContext)
			if u == nil {
				u = "unknown"
			}
			url = u.(string)
		}

		statusStr := strconv.Itoa(status)
		p.reqDur.WithLabelValues(statusStr, c.Request.Method, url).Observe(elapsed)
		p.reqCnt.WithLabelValues(statusStr, c.Request.Method, p.RequestCounterHostLabelMappingFunc(c), url).Inc()
		p.reqSz.WithLabelValues(statusStr, c.Request.Method, url).Observe(float64(reqSz))

		resSz := float64(c.Writer.Size())
		p.resSz.WithLabelValues(statusStr, c.Request.Method, url).Observe(resSz)
	}
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
