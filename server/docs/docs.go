package docs

import (
	"encoding/json"
	"strings"

	"github.com/swaggo/swag"
)

const docTemplate = `{
  "swagger": "2.0",
  "info": {
    "description": "Ingest and query home telemetry metrics.",
    "title": "Home Telemetry API",
    "version": "0.1.0"
  },
  "basePath": "/api/v1",
  "schemes": [
    "https"
  ],
  "paths": {
    "/ingest": {
      "post": {
        "tags": ["ingest"],
        "summary": "Ingest metrics",
        "description": "Accepts a single metrics payload.",
        "consumes": ["application/json"],
        "produces": ["application/json"],
        "parameters": [
          {
            "in": "body",
            "name": "payload",
            "required": true,
            "schema": {"$ref": "#/definitions/types.IngestPayload"}
          }
        ],
        "responses": {
          "202": {"description": "Accepted"},
          "400": {"description": "Bad Request"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Internal Server Error"}
        },
        "security": [{"BearerAuth": []}]
      }
    },
    "/nodes": {
      "get": {
        "tags": ["nodes"],
        "summary": "List nodes",
        "produces": ["application/json"],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {"type": "array", "items": {"$ref": "#/definitions/store.Node"}}
          },
          "401": {"description": "Unauthorized"},
          "500": {"description": "Internal Server Error"}
        },
        "security": [{"BearerAuth": []}]
      }
    },
    "/metrics": {
      "get": {
        "tags": ["metrics"],
        "summary": "Query metrics",
        "produces": ["application/json"],
        "parameters": [
          {"name": "node_id", "in": "query", "required": true, "type": "string"},
          {"name": "metric", "in": "query", "required": false, "type": "string"},
          {"name": "from", "in": "query", "required": false, "type": "string"},
          {"name": "to", "in": "query", "required": false, "type": "string"},
          {"name": "limit", "in": "query", "required": false, "type": "integer"}
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "properties": {
                "series": {"type": "array", "items": {"$ref": "#/definitions/types.MetricRow"}}
              }
            }
          },
          "400": {"description": "Bad Request"},
          "401": {"description": "Unauthorized"}
        },
        "security": [{"BearerAuth": []}]
      }
    }
  },
  "securityDefinitions": {
    "BearerAuth": {
      "type": "apiKey",
      "name": "Authorization",
      "in": "header"
    }
  },
  "definitions": {
    "types.CPUMetrics": {
      "type": "object",
      "properties": {
        "temp_c": {"type": "number"},
        "usage_pct": {"type": "number"},
        "load_avg": {"type": "number"},
        "freq_mhz": {"type": "number"},
        "cores": {"type": "integer"}
      }
    },
    "types.GPUMetrics": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "temp_c": {"type": "number"},
        "usage_pct": {"type": "number"},
        "mem_used_mb": {"type": "number"},
        "power_w": {"type": "number"}
      }
    },
    "types.IngestPayload": {
      "type": "object",
      "properties": {
        "node_id": {"type": "string"},
        "timestamp": {"type": "string"},
        "cpu": {"$ref": "#/definitions/types.CPUMetrics"},
        "gpus": {"type": "array", "items": {"$ref": "#/definitions/types.GPUMetrics"}},
        "tags": {"type": "object", "additionalProperties": {"type": "string"}}
      }
    },
    "store.Node": {
      "type": "object",
      "properties": {
        "id": {"type": "string"},
        "name": {"type": "string"},
        "last_seen": {"type": "string"},
        "meta": {"type": "object", "additionalProperties": {"type": "string"}}
      }
    },
    "types.MetricRow": {
      "type": "object",
      "properties": {
        "time": {"type": "string"},
        "metric": {"type": "string"},
        "value": {"type": "number"},
        "labels": {"type": "object", "additionalProperties": {"type": "string"}}
      }
    }
  }
}
`

var SwaggerInfo = &swaggerInfo{
	Version:     "0.1.0",
	Host:        "",
	BasePath:    "/api/v1",
	Schemes:     []string{"https"},
	Title:       "Home Telemetry API",
	Description: "Ingest and query home telemetry metrics.",
}

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

func (s *swaggerInfo) ReadDoc() string {
	replacer := strings.NewReplacer(
		"{{.Title}}", s.Title,
		"{{.Description}}", s.Description,
		"{{.Version}}", s.Version,
		"{{.Host}}", s.Host,
		"{{.BasePath}}", s.BasePath,
		"{{.Schemes}}", schemes(s.Schemes),
	)

	return replacer.Replace(docTemplate)
}

func schemes(s []string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func init() {
	swag.Register(swag.Name, SwaggerInfo)
}
