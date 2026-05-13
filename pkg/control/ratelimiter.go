package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

// per-key token bucket rate limiter for xApps
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rps      float64
	burst    int
	enabled  bool
}

func NewRateLimiter(rps float64, burst int, enabled bool) *RateLimiter {
	if rps <= 0 {
		rps = 100.0
	}
	if burst <= 0 {
		burst = 20
	}
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
		enabled:  enabled,
	}
}

func (r *RateLimiter) getLimiter(key string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	l, ok := r.limiters[key]
	if !ok {
		l = rate.NewLimiter(rate.Limit(r.rps), r.burst)
		r.limiters[key] = l
	}
	return l
}

func (r *RateLimiter) Allow(key string) bool {
	if !r.enabled {
		return true
	}
	if key == "" {
		return true
	}
	return r.getLimiter(key).Allow()
}

// Wrapper that lets us add rate limiting to the normal RIC handlers
type RateLimiterWrapper struct {
	limiter    *RateLimiter
	extractKey func(*http.Request) (string, error)
}

func NewRateLimiterWrapper(rl *RateLimiter, extractor func(*http.Request) (string, error)) *RateLimiterWrapper {
	return &RateLimiterWrapper{limiter: rl, extractKey: extractor}
}

// Wraps InjectRoute handlers with rate limiting logic
func (w *RateLimiterWrapper) WrapHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		key, err := w.extractKey(r)
		if err != nil {
			xapp.Logger.Error("RateLimiter: failed to extract xApp service name: %v", err)
			http.Error(rw, "Unable to identify xApp", http.StatusBadRequest)
			return
		}

		routeKey := fmt.Sprintf("%s|%s", key, r.URL.Path)
		if !w.limiter.Allow(routeKey) {
			xapp.Logger.Warn("RateLimiter: rate limit exceeded for %s", routeKey)
			http.Error(rw, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		handler(rw, r)
	}
}

// Wraps subscription handler - maintains original signature: func(params interface{}) (*models.SubscriptionResponse, int)
func (w *RateLimiterWrapper) WrapRESTSubscriptionHandler(
	handler func(params interface{}) (*models.SubscriptionResponse, int),
) func(params interface{}) (*models.SubscriptionResponse, int) {
	return func(params interface{}) (*models.SubscriptionResponse, int) {
		// Extract xApp service name from params
		var xAppServiceName string
		if p, ok := params.(*models.SubscriptionParams); ok && p.ClientEndpoint != nil {
			xAppServiceName = p.ClientEndpoint.Host
		}

		routeKey := fmt.Sprintf("%s|subscription", xAppServiceName)
		if !w.limiter.Allow(routeKey) {
			xapp.Logger.Warn("RateLimiter: rate limit exceeded for %s", routeKey)
			// Return rate limit error
			return nil, 429 // HTTP 429 Too Many Requests
		}

		// Call the original handler
		return handler(params)
	}
}

// Wraps subscription delete handler - maintains original signature: func(restSubId string) int
func (w *RateLimiterWrapper) WrapRESTSubscriptionDeleteHandler(
	handler func(restSubId string) int,
) func(restSubId string) int {
	return func(restSubId string) int {
		// For delete operations, we can use restSubId as the key or extract from context
		routeKey := fmt.Sprintf("%s|subscription_delete", restSubId)
		if !w.limiter.Allow(routeKey) {
			xapp.Logger.Warn("RateLimiter: rate limit exceeded for %s", routeKey)
			// Return rate limit error
			return 429 // HTTP 429 Too Many Requests
		}

		// Call the original handler
		return handler(restSubId)
	}
}

// Extracts xApp service name from JSON body and restores the body for downstream handlers.
func ExtractXappServiceName(r *http.Request) (string, error) {
	if r == nil {
		return "", fmt.Errorf("nil request")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if len(body) == 0 {
		return "", nil
	}

	var params models.SubscriptionParams
	if err := json.Unmarshal(body, &params); err != nil {
		return "", fmt.Errorf("unmarshal body: %w", err)
	}

	if params.ClientEndpoint == nil || params.ClientEndpoint.Host == "" {
		return "", fmt.Errorf("ClientEndpoint.Host not found")
	}

	return params.ClientEndpoint.Host, nil
}

// Reads limiter config (enabled, RPS, burst) from viper.
func ReadRateLimiterConfig() (float64, int, bool) {
	enabled := viper.GetBool("ratelimiter.enabled")
	rps := viper.GetFloat64("ratelimiter.requests_per_second")
	burst := viper.GetInt("ratelimiter.burst")

	if rps <= 0 {
		rps = 100.0
	}
	if burst <= 0 {
		burst = 20
	}

	return rps, burst, enabled
}

func NewDefaultRateLimiterWrapper() *RateLimiterWrapper {
	rps, burst, enabled := ReadRateLimiterConfig()
	rl := NewRateLimiter(rps, burst, enabled)
	return NewRateLimiterWrapper(rl, ExtractXappServiceName)
}
