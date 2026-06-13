package docs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var specYAML []byte

// Handler serves the OpenAPI spec and a Swagger UI page.
func Handler() http.Handler {
	mux := http.NewServeMux()

	// Serve the raw spec
	mux.HandleFunc("GET /api/v1/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Write(specYAML)
	})

	// Serve Swagger UI page
	mux.HandleFunc("GET /api/v1/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(swaggerHTML))
	})

	return mux
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>VoxMeet API</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/api/v1/docs/openapi.yaml",
      dom_id: "#swagger-ui",
    });
  </script>
</body>
</html>`
