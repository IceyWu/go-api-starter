package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ScalarHTML returns the Scalar API documentation HTML
const ScalarHTML = `<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
    <script id="api-reference" data-url="/swagger/doc.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

// DocsHandler serves the Scalar API documentation UI
func DocsHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, ScalarHTML)
}
