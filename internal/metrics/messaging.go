package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	msgPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ppo_messages_published_total",
			Help: "Published messages by queue",
		},
		[]string{"queue"},
	)
	msgConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ppo_messages_consumed_total",
			Help: "Consumed messages by queue",
		},
		[]string{"queue", "status"},
	)
	msgDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ppo_message_process_seconds",
			Help:    "Message processing duration by queue",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"queue", "status"},
	)
)

func IncPublished(queue string) {
	msgPublished.WithLabelValues(queue).Inc()
}

func ObserveConsume(queue, status string, dur time.Duration) {
	msgConsumed.WithLabelValues(queue, status).Inc()
	msgDuration.WithLabelValues(queue, status).Observe(dur.Seconds())
}
