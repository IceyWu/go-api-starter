package llmstxt

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// Handler serves llms.txt and llms-full.txt endpoints
type Handler struct {
	specJSON string
	cfg      Config

	once     sync.Once
	spec     *SwaggerSpec
	parseErr error
}

// NewHandler creates a new LLMs txt handler
func NewHandler(swaggerJSON string, cfg Config) *Handler {
	return &Handler{
		specJSON: swaggerJSON,
		cfg:      cfg,
	}
}

func (h *Handler) ensureParsed() {
	h.once.Do(func() {
		h.spec, h.parseErr = ParseSwagger([]byte(h.specJSON))
	})
}

// LLMsTxt serves GET /llms.txt
func (h *Handler) LLMsTxt(c *gin.Context) {
	h.ensureParsed()
	if h.parseErr != nil {
		c.String(http.StatusInternalServerError, "Failed to parse swagger spec")
		return
	}
	cfg := h.resolveConfig(c)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, GenerateLLMsTxt(h.spec, cfg))
}

// LLMsFullTxt serves GET /llms-full.txt
func (h *Handler) LLMsFullTxt(c *gin.Context) {
	h.ensureParsed()
	if h.parseErr != nil {
		c.String(http.StatusInternalServerError, "Failed to parse swagger spec")
		return
	}
	cfg := h.resolveConfig(c)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, GenerateLLMsFullTxt(h.spec, cfg))
}

// resolveConfig returns a Config with BaseURL resolved from the request if not set
func (h *Handler) resolveConfig(c *gin.Context) Config {
	if h.cfg.BaseURL != "" {
		return h.cfg
	}
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return Config{
		BaseURL: scheme + "://" + c.Request.Host,
	}
}

// RegisterRoutes registers llms.txt routes on the given router group or engine
func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/llms.txt", h.LLMsTxt)
	r.GET("/llms-full.txt", h.LLMsFullTxt)
}
