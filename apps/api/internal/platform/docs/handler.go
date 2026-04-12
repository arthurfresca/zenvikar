package docs

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed specs/*.yaml specs/domains/*.yaml
var openAPIFS embed.FS

const swaggerUIHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Zenvikar API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "%s",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

// OpenAPIHandler serves the OpenAPI specification YAML file.
func OpenAPIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spec, err := openAPIFS.ReadFile("specs/openapi.yaml")
		if err != nil {
			http.Error(w, "failed to load openapi specification", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(spec)
	}
}

// DomainSpecsHandler serves additional per-domain OpenAPI files used by $ref.
func DomainSpecsHandler() http.Handler {
	sub, err := fs.Sub(openAPIFS, "specs/domains")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "failed to load domain specs", http.StatusInternalServerError)
		})
	}

	return http.StripPrefix("/swagger/domains/", http.FileServer(http.FS(sub)))
}

// SwaggerUIHandler serves a lightweight Swagger UI page.
func SwaggerUIHandler(specPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(swaggerUIHTML, specPath)))
	}
}
