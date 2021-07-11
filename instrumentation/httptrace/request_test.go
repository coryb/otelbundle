package httptrace_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coryb/otelbundle/instrumentation/httptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type spanProcessor struct {
	Ended chan sdktrace.ReadOnlySpan
}

var _ sdktrace.SpanProcessor = (*spanProcessor)(nil)

func (sp *spanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	go func() {
		sp.Ended <- s
	}()
}
func (sp *spanProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {}
func (sp *spanProcessor) Shutdown(ctx context.Context) error                       { return nil }
func (sp *spanProcessor) ForceFlush(ctx context.Context) error                     { return nil }

var testProcessor = &spanProcessor{
	Ended: make(chan sdktrace.ReadOnlySpan),
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

func init() {
	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithIDGenerator(&testIDGenerator{}),
			sdktrace.WithSpanProcessor(testProcessor),
		),
	)
}

func ExampleTransport() {
	client := &http.Client{
		Transport: otelhttp.NewTransport(
			httptrace.Transport(http.DefaultTransport),
			otelhttp.WithSpanNameFormatter(func(_ string, req *http.Request) string {
				return fmt.Sprintf("%s %s", req.Method, req.URL.Path)
			}),
		),
	}
	resp, err := client.Get("https://example.com")
	if err != nil {
		panic(fmt.Sprintf("Failed to GET https://example.com %q", err))
	}
	resp.Body.Close()

	// global span processor for testing
	testSpan := <-testProcessor.Ended
	for _, ev := range testSpan.Events() {
		fmt.Printf("Got Event: %s\n", ev.Name)
	}

	// Output:
	// Got Event: Connecting
	// Got Event: DNS Start
	// Got Event: DNS Done
	// Got Event: Connect Start
	// Got Event: Connect Done
	// Got Event: TLS Handshake Start
	// Got Event: TLS Handshake Done
	// Got Event: Connected
	// Got Event: Wrote Headers
	// Got Event: Wrote Request
	// Got Event: First Response Byte
}
