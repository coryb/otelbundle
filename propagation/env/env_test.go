package env_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/coryb/otelbundle/propagation/env"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithIDGenerator(&testIDGenerator{}),
		),
	)
}

type testIDGenerator struct{}

var _ sdktrace.IDGenerator = (*testIDGenerator)(nil)

func (g *testIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	traceID, _ := trace.TraceIDFromHex("60d19e9e9abf2197c1d6d8f93e28ee2a")
	spanID, _ := trace.SpanIDFromHex("a028bd951229a46f")
	return traceID, spanID
}

func (g *testIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	spanID, _ := trace.SpanIDFromHex("a028bd951229a46f")
	return spanID
}

func TestInject(t *testing.T) {
	tracer := otel.Tracer("example")
	ctx, span := tracer.Start(context.Background(), "testing")
	defer span.End()

	input := []string{"PATH=/usr/bin:/bin"}

	otel.SetTextMapPropagator(propagation.TraceContext{})
	got := env.Inject(ctx, input)
	require.Equal(t, []string{
		"PATH=/usr/bin:/bin",
		"TRACESTATE=",
		"TRACEPARENT=00-60d19e9e9abf2197c1d6d8f93e28ee2a-a028bd951229a46f-01",
	}, got)

	otel.SetTextMapPropagator(b3.B3{InjectEncoding: b3.B3MultipleHeader})
	got = env.Inject(ctx, input)
	require.Equal(t, []string{
		"PATH=/usr/bin:/bin",
		"X_B3_TRACEID=60d19e9e9abf2197c1d6d8f93e28ee2a",
		"X_B3_SPANID=a028bd951229a46f",
		"X_B3_SAMPLED=1",
	}, got)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			b3.B3{InjectEncoding: b3.B3MultipleHeader},
			propagation.TraceContext{},
		),
	)
	got = env.Inject(ctx, input)
	require.Equal(t, []string{
		"PATH=/usr/bin:/bin",
		"X_B3_TRACEID=60d19e9e9abf2197c1d6d8f93e28ee2a",
		"X_B3_SPANID=a028bd951229a46f",
		"X_B3_SAMPLED=1",
		"TRACESTATE=",
		"TRACEPARENT=00-60d19e9e9abf2197c1d6d8f93e28ee2a-a028bd951229a46f-01",
	}, got)

	// verify we update rather than append
	input = []string{
		"PATH=/usr/bin:/bin",
		"TRACESTATE=origTraceState",
		"TRACEPARENT=origTraceParent",
		"X_B3_TRACEID=origTraceID",
		"X_B3_SPANID=origSpanID",
		"X_B3_SAMPLED=origSampled",
		"TERM=xterm",
	}

	otel.SetTextMapPropagator(propagation.TraceContext{})
	got = env.Inject(ctx, input)
	require.Equal(t, []string{
		"PATH=/usr/bin:/bin",
		"TRACESTATE=",
		"TRACEPARENT=00-60d19e9e9abf2197c1d6d8f93e28ee2a-a028bd951229a46f-01",
		// these left unchanged since we are not using b3 propagator
		"X_B3_TRACEID=origTraceID",
		"X_B3_SPANID=origSpanID",
		"X_B3_SAMPLED=origSampled",
		"TERM=xterm",
	}, got)

	otel.SetTextMapPropagator(b3.B3{InjectEncoding: b3.B3MultipleHeader})
	got = env.Inject(ctx, input)
	require.Equal(t, []string{
		"PATH=/usr/bin:/bin",
		// these left unchanged since we are not using trace context propagator
		"TRACESTATE=origTraceState",
		"TRACEPARENT=origTraceParent",
		"X_B3_TRACEID=60d19e9e9abf2197c1d6d8f93e28ee2a",
		"X_B3_SPANID=a028bd951229a46f",
		"X_B3_SAMPLED=1",
		"TERM=xterm",
	}, got)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			b3.B3{InjectEncoding: b3.B3MultipleHeader},
			propagation.TraceContext{},
		),
	)
	got = env.Inject(ctx, input)
	require.Equal(t, []string{
		"PATH=/usr/bin:/bin",
		"TRACESTATE=",
		"TRACEPARENT=00-60d19e9e9abf2197c1d6d8f93e28ee2a-a028bd951229a46f-01",
		"X_B3_TRACEID=60d19e9e9abf2197c1d6d8f93e28ee2a",
		"X_B3_SPANID=a028bd951229a46f",
		"X_B3_SAMPLED=1",
		"TERM=xterm",
	}, got)
}

func ExampleInject() {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	tracer := otel.Tracer("example")
	ctx, span := tracer.Start(context.Background(), "testing")
	defer span.End()

	cmd := exec.Command("env")
	cmd.Stdout = os.Stdout
	cmd.Env = env.Inject(ctx, []string{"PATH=/usr/bin:/bin"})
	cmd.Run()

	// Output:
	// PATH=/usr/bin:/bin
	// TRACESTATE=
	// TRACEPARENT=00-60d19e9e9abf2197c1d6d8f93e28ee2a-a028bd951229a46f-01
}

func ExampleContextWithRemoteSpanContext() {
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// these would normally be imported from environment
	os.Setenv("TRACESTATE", "")
	os.Setenv("TRACEPARENT", "00-60d19e9e9abf2197c1d6d8f93e28ee2a-a028bd951229a46f-01")

	// extract span context from environ
	ctx := env.ContextWithRemoteSpanContext(context.Background(), os.Environ())
	span := trace.SpanFromContext(ctx)

	fmt.Printf("TraceID: %s\n", span.SpanContext().TraceID().String())
	fmt.Printf("SpanID: %s\n", span.SpanContext().SpanID().String())
	fmt.Printf("Is Sampled: %t\n", span.SpanContext().IsSampled())

	// Output:
	// TraceID: 60d19e9e9abf2197c1d6d8f93e28ee2a
	// SpanID: a028bd951229a46f
	// Is Sampled: true
}

func ExampleInject_b3() {
	otel.SetTextMapPropagator(b3.B3{InjectEncoding: b3.B3MultipleHeader})
	tracer := otel.Tracer("example")
	ctx, span := tracer.Start(context.Background(), "testing")
	defer span.End()

	cmd := exec.Command("env")
	cmd.Stdout = os.Stdout
	cmd.Env = env.Inject(ctx, []string{"PATH=/usr/bin:/bin"})
	cmd.Run()

	// Output:
	// PATH=/usr/bin:/bin
	// X_B3_TRACEID=60d19e9e9abf2197c1d6d8f93e28ee2a
	// X_B3_SPANID=a028bd951229a46f
	// X_B3_SAMPLED=1
}

func ExampleContextWithRemoteSpanContext_b3() {
	otel.SetTextMapPropagator(b3.B3{InjectEncoding: b3.B3MultipleHeader})

	// these would normally be imported from environment
	os.Setenv("X_B3_TRACEID", "60d19e9e9abf2197c1d6d8f93e28ee2a")
	os.Setenv("X_B3_SPANID", "a028bd951229a46f")
	os.Setenv("X_B3_SAMPLED", "1")

	// extract span context from environ
	ctx := env.ContextWithRemoteSpanContext(context.Background(), os.Environ())
	span := trace.SpanFromContext(ctx)

	fmt.Printf("TraceID: %s\n", span.SpanContext().TraceID().String())
	fmt.Printf("SpanID: %s\n", span.SpanContext().SpanID().String())
	fmt.Printf("Is Sampled: %t\n", span.SpanContext().IsSampled())

	// Output:
	// TraceID: 60d19e9e9abf2197c1d6d8f93e28ee2a
	// SpanID: a028bd951229a46f
	// Is Sampled: true
}
