package env

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Carrier is a propagation.TextMapCarrier that is used to extract or
// inject tracing environment variables. This is used with a
// propagation.TextMapPropagator
type Carrier struct {
	Env []string
}

var _ propagation.TextMapCarrier = (*Carrier)(nil)

func toKey(key string) string {
	key = strings.ToUpper(key)
	key = strings.ReplaceAll(key, "-", "_")
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			return r
		}
		return -1
	}, key)
}

func (c *Carrier) Get(key string) string {
	if c == nil {
		return ""
	}
	key = toKey(key)
	for _, e := range c.Env {
		if strings.HasPrefix(e, key+"=") {
			return strings.TrimPrefix(e, key+"=")
		}
	}
	return ""
}

func (c *Carrier) Set(key, value string) {
	if c == nil {
		return
	}
	key = toKey(key)
	for i, e := range c.Env {
		if strings.HasPrefix(e, key+"=") {
			// don't directly update the slice so we don't modify the slice
			// passed in
			newEnv := make([]string, len(c.Env))
			copy(newEnv, c.Env)
			c.Env = append(newEnv[:i], append([]string{fmt.Sprintf("%s=%s", key, value)}, newEnv[i+1:]...)...)
			return
		}
	}
	c.Env = append(c.Env, fmt.Sprintf("%s=%s", key, value))
}

func (c *Carrier) Keys() []string {
	if c == nil {
		return nil
	}
	keys := make([]string, len(c.Env))
	var parts []string
	for _, e := range c.Env {
		parts = strings.SplitN(e, "=", 2)
		keys = append(keys, parts[0])
	}
	return keys
}

// Inject will add add any necessary environment variables for the span
// found in the Context.  If environment variables are already present
// in `environ` then they will be updated.  If no variables are found the
// new ones will be appended.  The new environment will be returned, `environ`
// will never be modified.
func Inject(ctx context.Context, environ []string) []string {
	carrier := &Carrier{Env: environ}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier.Env
}

// ContextWithRemoteSpanContext extracts a remote SpanContext from the
// environment and will set the new SpanContext on the returned Context
func ContextWithRemoteSpanContext(ctx context.Context, environ []string) context.Context {
	carrier := &Carrier{Env: environ}
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
