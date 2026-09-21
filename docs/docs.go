// Package docs serves the checked-in OpenAPI document.
package docs

import (
	_ "embed"
)

//go:embed swagger.json
var swaggerJSON []byte

// ReadDoc returns the checked-in OpenAPI document.
func ReadDoc() string {
	return string(swaggerJSON)
}
