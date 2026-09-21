package handler

import (
	"fmt"
	"net/http"

	"go-api-starter/internal/transport"

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
      // Scalar renders OpenAPI Markdown inside web components, so handle these
      // two document links at the composed event path to open them in a new tab.
      document.addEventListener('click', function (event) {
        var link = event.composedPath().find(function (node) {
          return node instanceof HTMLAnchorElement;
        });
        if (!link) return;

        var path = new URL(link.href, window.location.href).pathname;
        if (!path.endsWith('/llms.txt') && !path.endsWith('/llms-full.txt')) return;

        event.preventDefault();
        window.open(link.href, '_blank', 'noopener,noreferrer');
      }, true);

      Scalar.createApiReference('#app', {
        url: '%s/openapi.json',
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
func DocsHandler(c *transport.Context) {
	cfg := config.GetConfig()
	appName := "Go API Starter"
	favicon := "/favicon.ico"
	basePath := ""
	if cfg != nil && cfg.App.Name != "" {
		appName = cfg.App.Name
	}
	if cfg != nil {
		basePath = cfg.Server.BasePath
	}

	html := fmt.Sprintf(docsTemplate, appName, favicon, basePath, favicon, appName)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
