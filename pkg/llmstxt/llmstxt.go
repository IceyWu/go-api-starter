package llmstxt

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Config holds the configuration for generating llms.txt
type Config struct {
	BaseURL string // e.g. "http://localhost:9527"
}

// SwaggerSpec represents the Swagger 2.0 spec structure
type SwaggerSpec struct {
	Info        InfoBlock                      `json:"info"`
	Paths       map[string]PathItem            `json:"paths"`
	Definitions map[string]SchemaObject        `json:"definitions"`
}

type InfoBlock struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type PathItem map[string]Operation // method -> operation

type Operation struct {
	Summary     string                  `json:"summary"`
	Description string                  `json:"description"`
	Tags        []string                `json:"tags"`
	Parameters  []Parameter             `json:"parameters"`
	Responses   map[string]ResponseItem `json:"responses"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type ResponseItem struct {
	Description string       `json:"description"`
	Schema      *SchemaRef   `json:"schema"`
}

type SchemaRef struct {
	Ref        string            `json:"$ref"`
	Type       string            `json:"type"`
	Properties map[string]*SchemaRef `json:"properties"`
	AllOf      []SchemaRef       `json:"allOf"`
	Example    interface{}       `json:"example"`
}

type SchemaObject struct {
	Type       string                `json:"type"`
	Properties map[string]*Property  `json:"properties"`
	Required   []string              `json:"required"`
}

type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Example     interface{} `json:"example"`
	Ref         string      `json:"$ref"`
	Enum        []interface{} `json:"enum"`
}

// ParseSwagger parses swagger JSON string into SwaggerSpec
func ParseSwagger(data []byte) (*SwaggerSpec, error) {
	var spec SwaggerSpec
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	// Use a more lenient approach
	if err := json.Unmarshal(data, &spec); err != nil {
		// Try with string replacement for control chars
		cleaned := strings.Map(func(r rune) rune {
			if r < 0x20 && r != '\n' && r != '\r' && r != '\t' {
				return ' '
			}
			return r
		}, string(data))
		if err2 := json.Unmarshal([]byte(cleaned), &spec); err2 != nil {
			return nil, err2
		}
	}
	_ = decoder
	return &spec, nil
}

// generateExample builds a JSON example from a response schema
func generateExample(schema *SchemaRef, defs map[string]SchemaObject, depth int) interface{} {
	if depth > 5 {
		return nil
	}

	if schema == nil {
		return nil
	}

	// Handle $ref
	if schema.Ref != "" {
		refName := strings.TrimPrefix(schema.Ref, "#/definitions/")
		if def, ok := defs[refName]; ok {
			return generateExampleFromDef(def, defs, depth+1)
		}
		return nil
	}

	// Handle allOf (common in swaggo output)
	if len(schema.AllOf) > 0 {
		merged := make(map[string]interface{})
		for _, item := range schema.AllOf {
			result := generateExample(&item, defs, depth+1)
			if m, ok := result.(map[string]interface{}); ok {
				for k, v := range m {
					merged[k] = v
				}
			}
		}
		return merged
	}

	// Handle inline properties
	if schema.Properties != nil {
		obj := make(map[string]interface{})
		for name, prop := range schema.Properties {
			obj[name] = generateExample(prop, defs, depth+1)
		}
		return obj
	}

	// Handle example value
	if schema.Example != nil {
		return schema.Example
	}

	// Default by type
	switch schema.Type {
	case "string":
		return "string"
	case "integer":
		return 0
	case "number":
		return 0.0
	case "boolean":
		return false
	case "array":
		return []interface{}{}
	}

	return nil
}

func generateExampleFromDef(def SchemaObject, defs map[string]SchemaObject, depth int) interface{} {
	if depth > 5 {
		return nil
	}
	obj := make(map[string]interface{})
	for name, prop := range def.Properties {
		if prop.Ref != "" {
			refName := strings.TrimPrefix(prop.Ref, "#/definitions/")
			if refDef, ok := defs[refName]; ok {
				obj[name] = generateExampleFromDef(refDef, defs, depth+1)
			}
		} else if prop.Example != nil {
			obj[name] = prop.Example
		} else {
			switch prop.Type {
			case "string":
				obj[name] = "string"
			case "integer":
				obj[name] = 0
			case "number":
				obj[name] = 0.0
			case "boolean":
				obj[name] = false
			case "array":
				obj[name] = []interface{}{}
			default:
				obj[name] = nil
			}
		}
	}
	return obj
}

// GenerateLLMsTxt generates the concise llms.txt content
func GenerateLLMsTxt(spec *SwaggerSpec, cfg Config) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", spec.Info.Title))
	sb.WriteString(fmt.Sprintf("> %s\n\n", spec.Info.Description))
	sb.WriteString(fmt.Sprintf("Base URL: %s\n\n", cfg.BaseURL))

	sb.WriteString("## 文档\n\n")
	sb.WriteString(fmt.Sprintf("- [OpenAPI Spec](%s/swagger/doc.json): 机器可读的完整接口定义（推荐 AI 直接使用）\n", cfg.BaseURL))
	sb.WriteString(fmt.Sprintf("- [接口文档](%s/docs): 可视化接口文档\n", cfg.BaseURL))
	sb.WriteString(fmt.Sprintf("- [完整文档](%s/llms-full.txt): Markdown 格式的详细接口说明\n\n", cfg.BaseURL))

	sb.WriteString("## 接口概览\n\n")

	paths := sortedPaths(spec.Paths)
	for _, path := range paths {
		methods := spec.Paths[path]
		for _, method := range sortedMethods(methods) {
			op := methods[method]
			sb.WriteString(fmt.Sprintf("- %s %s: %s\n", strings.ToUpper(method), path, op.Summary))
		}
	}
	sb.WriteString("\n")

	return sb.String()
}

// GenerateLLMsFullTxt generates the detailed llms-full.txt content
func GenerateLLMsFullTxt(spec *SwaggerSpec, cfg Config) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s - 完整接口文档\n\n", spec.Info.Title))
	sb.WriteString(fmt.Sprintf("> %s\n\n", spec.Info.Description))
	sb.WriteString(fmt.Sprintf("Base URL: %s\n\n", cfg.BaseURL))
	sb.WriteString("---\n\n")

	paths := sortedPaths(spec.Paths)
	for _, path := range paths {
		methods := spec.Paths[path]
		for _, method := range sortedMethods(methods) {
			op := methods[method]

			sb.WriteString(fmt.Sprintf("## %s %s\n\n", strings.ToUpper(method), path))
			if op.Summary != "" {
				sb.WriteString(fmt.Sprintf("**%s**\n\n", op.Summary))
			}
			if op.Description != "" {
				sb.WriteString(op.Description + "\n\n")
			}
			if len(op.Tags) > 0 {
				sb.WriteString(fmt.Sprintf("标签: %s\n\n", strings.Join(op.Tags, ", ")))
			}

			// Parameters
			if len(op.Parameters) > 0 {
				sb.WriteString("### 参数\n\n")
				sb.WriteString("| 字段 | 位置 | 类型 | 必填 | 说明 |\n")
				sb.WriteString("|------|------|------|------|------|\n")
				for _, p := range op.Parameters {
					required := "否"
					if p.Required {
						required = "是"
					}
					pType := p.Type
					if pType == "" {
						pType = "object"
					}
					desc := p.Description
					if desc == "" {
						desc = p.Name
					}
					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", p.Name, p.In, pType, required, desc))
				}
				sb.WriteString("\n")
			}

			// Responses with examples
			if len(op.Responses) > 0 {
				sb.WriteString("### 响应\n\n")
				for code, resp := range op.Responses {
					sb.WriteString(fmt.Sprintf("- **%s**: %s\n", code, resp.Description))

					// Generate example if schema is available
					if resp.Schema != nil {
						example := generateExample(resp.Schema, spec.Definitions, 0)
						if example != nil {
							exJSON, err := json.MarshalIndent(example, "  ", "  ")
							if err == nil {
								sb.WriteString(fmt.Sprintf("\n  ```json\n  %s\n  ```\n\n", string(exJSON)))
							}
						}
					}
				}
				sb.WriteString("\n")
			}

			sb.WriteString("---\n\n")
		}
	}

	return sb.String()
}

func sortedPaths(paths map[string]PathItem) []string {
	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedMethods(item PathItem) []string {
	order := map[string]int{"get": 0, "post": 1, "put": 2, "patch": 3, "delete": 4}
	keys := make([]string, 0, len(item))
	for k := range item {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		oi, ok1 := order[keys[i]]
		oj, ok2 := order[keys[j]]
		if !ok1 {
			oi = 99
		}
		if !ok2 {
			oj = 99
		}
		return oi < oj
	})
	return keys
}
