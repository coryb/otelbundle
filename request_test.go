package otelhttptrace_test

import (
	"fmt"
	"net/http"

	"github.com/coryb/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func ExampleTransport() {
	client := &http.Client{
		Transport: otelhttp.NewTransport(
			otelhttptrace.Transport(http.DefaultTransport),
			otelhttp.WithSpanNameFormatter(func(_ string, req *http.Request) string {
				return fmt.Sprintf("%s %s", req.Method, req.URL.Path)
			}),
		),
	}
	resp, err := client.Get("https://github.com")
	if err != nil {
		panic(fmt.Sprintf("Failed to GET https://github.com %q", err))
	}
	defer resp.Body.Close()
	fmt.Printf("Got %s", resp.Status)

	// Output:
	// Got 200 OK
}
