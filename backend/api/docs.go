// Package api serves the OpenAPI contract and an interactive reference rendered by Scalar.
//
// The contract (openapi.yaml) is written by hand and embedded in the binary, so the running
// server always documents itself. Keep it in step with the controllers: any change to a route,
// a request body or a response belongs in the same commit as the change to the YAML.
package api

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var spec []byte

// Spec returns the OpenAPI document.
func Spec() []byte { return spec }

// scalarVersion is pinned so the documentation page does not change under our feet, and the
// script carries a Subresource Integrity hash so a compromised CDN cannot run code on the API
// origin (where the refresh cookie is sent). The URL names the package's own dist file, not the
// jsDelivr "default file" alias, because jsDelivr re-minifies the alias. When bumping the version
// recompute the hash:
//
//	curl -sL https://cdn.jsdelivr.net/npm/@scalar/api-reference@VERSION/dist/browser/standalone.js \
//	  | openssl dgst -sha384 -binary | openssl base64 -A
const (
	scalarVersion   = "1.75.0"
	scalarIntegrity = "sha384-xgelKlDcc+BYp0O+ZYzFesLnWCgGpEbXVVZWNQT6WyZYygN01tbEMygnjPXb5OPI"
)

const referencePage = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>WebChat API reference</title>
  </head>
  <body>
    <script id="api-reference" data-url="/openapi.yaml"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@` + scalarVersion + `/dist/browser/standalone.js" integrity="` + scalarIntegrity + `" crossorigin="anonymous"></script>
  </body>
</html>
`

// RegisterDocs mounts GET /docs (the Scalar reference) and GET /openapi.yaml (the contract).
func RegisterDocs(router *gin.Engine) {
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", spec)
	})
	router.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(referencePage))
	})
}
