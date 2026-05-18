package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type openAPIDocument struct {
	OpenAPI    string                    `json:"openapi"`
	Info       openAPIInfo               `json:"info"`
	Servers    []openAPIServer           `json:"servers"`
	Paths      map[string]map[string]any `json:"paths"`
	Components openAPIComponents         `json:"components"`
}

type openAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type openAPIServer struct {
	URL string `json:"url"`
}

type openAPIComponents struct {
	Schemas map[string]any `json:"schemas"`
}

func registerSwaggerRoutes(router *gin.Engine) {
	router.GET("/swagger", swaggerIndexRedirect)
	router.GET("/swagger/doc.json", swaggerDocument)
	router.GET("/swagger/index.html", swaggerUI)
}

func swaggerIndexRedirect(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
}

func swaggerDocument(c *gin.Context) {
	c.JSON(http.StatusOK, newOpenAPIDocument())
}

func swaggerUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Marketplace Assistant API</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.ui = SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: "#swagger-ui"
      });
    </script>
  </body>
</html>`)
}

func newOpenAPIDocument() openAPIDocument {
	return openAPIDocument{
		OpenAPI: "3.0.3",
		Info: openAPIInfo{
			Title:       "Social Media Marketplace Assistant API",
			Description: "Private inventory and listing assistant API.",
			Version:     "0.1.0",
		},
		Servers: []openAPIServer{
			{URL: "/"},
		},
		Paths: map[string]map[string]any{
			"/healthz": {
				"get": map[string]any{
					"summary":     "Check API health",
					"operationId": "getHealth",
					"responses": map[string]any{
						"200": responseRef("HealthResponse"),
					},
				},
			},
			"/items": {
				"get": map[string]any{
					"summary":     "List inventory items",
					"operationId": "listItems",
					"parameters": []map[string]any{
						{
							"name":        "status",
							"in":          "query",
							"required":    false,
							"description": "Filter by inventory status.",
							"schema":      schemaRef("InventoryStatus"),
						},
					},
					"responses": map[string]any{
						"200": responseRef("ListItemsResponse"),
						"400": responseRef("ErrorResponse"),
					},
				},
				"post": map[string]any{
					"summary":     "Create an inventory item",
					"operationId": "createItem",
					"requestBody": requestBodyRef("CreateItemRequest"),
					"responses": map[string]any{
						"201": responseRef("ItemResponse"),
						"400": responseRef("ErrorResponse"),
					},
				},
			},
			"/items/{id}": {
				"get": map[string]any{
					"summary":     "Get an inventory item",
					"operationId": "getItem",
					"parameters":  []map[string]any{pathIDParameter()},
					"responses": map[string]any{
						"200": responseRef("ItemResponse"),
						"404": responseRef("ErrorResponse"),
					},
				},
				"patch": map[string]any{
					"summary":     "Update an inventory item",
					"operationId": "updateItem",
					"parameters":  []map[string]any{pathIDParameter()},
					"requestBody": requestBodyRef("UpdateItemRequest"),
					"responses": map[string]any{
						"200": responseRef("ItemResponse"),
						"400": responseRef("ErrorResponse"),
						"404": responseRef("ErrorResponse"),
					},
				},
				"delete": map[string]any{
					"summary":     "Archive an inventory item",
					"operationId": "archiveItem",
					"parameters":  []map[string]any{pathIDParameter()},
					"responses": map[string]any{
						"204": map[string]any{
							"description": "Item archived.",
						},
						"404": responseRef("ErrorResponse"),
					},
				},
			},
		},
		Components: openAPIComponents{
			Schemas: map[string]any{
				"HealthResponse": map[string]any{
					"type":     "object",
					"required": []string{"status", "service", "time"},
					"properties": map[string]any{
						"status":  stringSchema(),
						"service": stringSchema(),
						"time":    stringSchemaWithFormat("date-time"),
					},
				},
				"InventoryStatus": map[string]any{
					"type": "string",
					"enum": []string{"draft", "ready_to_list", "listed", "sold", "archived"},
				},
				"Currency": map[string]any{
					"type": "string",
					"enum": []string{"USD"},
				},
				"CreateItemRequest": itemRequestSchema(false),
				"UpdateItemRequest": itemRequestSchema(true),
				"ItemResponse": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":                            stringSchema(),
						"title":                         stringSchema(),
						"description":                   stringSchema(),
						"category":                      stringSchema(),
						"size":                          stringSchema(),
						"condition":                     stringSchema(),
						"original_purchase_price_cents": integerSchema(),
						"selling_price_cents":           integerSchema(),
						"currency":                      schemaRef("Currency"),
						"status":                        schemaRef("InventoryStatus"),
						"notes":                         stringSchema(),
						"created_at":                    stringSchemaWithFormat("date-time"),
						"updated_at":                    stringSchemaWithFormat("date-time"),
					},
				},
				"ListItemsResponse": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"items": map[string]any{
							"type":  "array",
							"items": schemaRef("ItemResponse"),
						},
					},
				},
				"ErrorResponse": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"error": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"code":    stringSchema(),
								"message": stringSchema(),
							},
						},
					},
				},
			},
		},
	}
}

func itemRequestSchema(partial bool) map[string]any {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title":                         stringSchema(),
			"description":                   stringSchema(),
			"category":                      stringSchema(),
			"size":                          stringSchema(),
			"condition":                     stringSchema(),
			"original_purchase_price_cents": integerSchema(),
			"selling_price_cents":           integerSchema(),
			"currency":                      schemaRef("Currency"),
			"notes":                         stringSchema(),
		},
	}
	if partial {
		schema["properties"].(map[string]any)["status"] = schemaRef("InventoryStatus")
		return schema
	}
	schema["required"] = []string{"title"}
	return schema
}

func pathIDParameter() map[string]any {
	return map[string]any{
		"name":     "id",
		"in":       "path",
		"required": true,
		"schema":   stringSchema(),
	}
}

func requestBodyRef(schema string) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": schemaRef(schema),
			},
		},
	}
}

func responseRef(schema string) map[string]any {
	return map[string]any{
		"description": "Response",
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": schemaRef(schema),
			},
		},
	}
}

func schemaRef(schema string) map[string]any {
	return map[string]any{
		"$ref": "#/components/schemas/" + schema,
	}
}

func stringSchema() map[string]any {
	return map[string]any{"type": "string"}
}

func stringSchemaWithFormat(format string) map[string]any {
	return map[string]any{
		"type":   "string",
		"format": format,
	}
}

func integerSchema() map[string]any {
	return map[string]any{
		"type":   "integer",
		"format": "int64",
	}
}
