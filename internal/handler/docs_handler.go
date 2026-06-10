package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-api-starter/internal/config"
)

const docsTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>%s - API Docs</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="icon" href="%s" />
</head>
<body>
    <div id="app"></div>
    <script src="https://fastly.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/swagger/doc.json',
        theme: 'elysiajs',
        darkMode: true,
        showSidebar: true,
        hideClientButton: false,
        showDeveloperTools: 'always',
        persistAuth: true,
        favicon: '%s',
        metaData: {
          title: '%s - API Docs',
        },
      })
    </script>
</body>
</html>`

// DocsHandler serves the Scalar API documentation UI.
// Logo and title are read from the OpenAPI info.title (set via swagger annotations).
// To display a logo in the sidebar, add x-logo to your OpenAPI info via swagger doc customization.
func DocsHandler(c *gin.Context) {
	cfg := config.GetConfig()
	appName := "Go API Starter"
	favicon := "/favicon.ico"
	if cfg != nil && cfg.App.Name != "" {
		appName = cfg.App.Name
	}

	html := fmt.Sprintf(docsTemplate, appName, favicon, favicon, appName)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
