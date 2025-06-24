package medication

import (
	"encoding/json"
	"io"
	"net/http"
)

// TODO: we should have some sort of API keys or user authorisation (typically JWT)
// for now we just accept X-Med-Owner header to test the functionality
func getOwner(r *http.Request) string {
	const defaultOwner = "default-owner"

	ownerHeader := r.Header.Get("X-Med-Owner")
	if ownerHeader == "" {
		return defaultOwner
	}
	return ownerHeader
}

func readJson(r *http.Request, v any) error {
	const maxJsonBytes = 10 * 1024 * 1024

	return json.NewDecoder(io.LimitReader(r.Body, maxJsonBytes)).Decode(v)
}
