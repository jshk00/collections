// nolint: revive
package main

import (
	"github.com/jshk00/collections/httpx"
)

func main() {
	httpx.NewCircuitBreaker(httpx.BreakerConfig{})
	// hc := httpx.New()
	// res, err := hc.Get("https://example.com").
	// 	SetHeaders(map[string]string{"Auth": "simple"}).
	// 	SetQueries(map[string]string{"id": "abcd"}).
	// 	RetryHook((&hooks.Retry{}).Hook).
	// 	Exec()
	// if err != nil {
	// 	panic(err)
	// }
	// defer res.Body.Close()
}
