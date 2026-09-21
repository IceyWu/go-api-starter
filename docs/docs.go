// Package docs serves the checked-in OpenAPI document.
package docs

import (
	_ "embed"
	"encoding/json"
)

//go:embed swagger.json
var swaggerJSON []byte

// ReadDoc returns the checked-in OpenAPI document.
func ReadDoc() string {
	return string(swaggerJSON)
}

// ReadDocForRequest returns the OpenAPI document with the public request host
// and configured base path. This keeps generated client examples usable from
// localhost, LAN, and reverse-proxy deployments without maintaining variants.
func ReadDocForRequest(host, basePath, scheme string) string {
	var document map[string]any
	if err := json.Unmarshal(swaggerJSON, &document); err != nil {
		return ReadDoc()
	}
	document["host"] = host
	document["basePath"] = basePath
	if scheme != "" {
		document["schemes"] = []string{scheme}
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return ReadDoc()
	}
	return string(encoded)
}
