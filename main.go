// nolint: revive
package main

import (
	"context"

	"collections/httpx"
	"collections/httpx/hooks"
)

func main() {
	hc := httpx.New(false)
	res, err := hc.Get("https://example.com").
		Context(context.Background()).
		Headers(map[string]string{"Auth": "simple"}).
		Queries(map[string]string{"id": "abcd"}).
		RetryHook((&hooks.RetryHook{}).Hook).
		Exec()
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
}
