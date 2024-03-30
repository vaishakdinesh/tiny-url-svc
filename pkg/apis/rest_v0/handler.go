package rest_v0

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vaishakdinesh/tiny-url-svc/types"
	v0 "github.com/vaishakdinesh/tiny-url-svc/types/api/rest/v0"
)

type handler struct {
	apiVersion string
	schema     types.OpenAPISchema
	svc        types.URLService
	l          *zap.Logger
}

func (h *handler) Register(s *types.Server) {
	sg := s.Group("/api/" + h.apiVersion)
	sg.Use(h.schema.ValidationMiddleware())
	v0.RegisterHandlers(sg, h)
}

func (h *handler) GenerateURL(ctx echo.Context) error {
	genURLReq := &v0.GenerateURLRequest{}
	err := json.NewDecoder(ctx.Request().Body).Decode(genURLReq)
	if err != nil {
		h.l.Error("failed to decode request", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, nil)
	}
	tinyURL, err := h.svc.GenerateTinyURL(ctx.Request().Context(), genURLReq.Url)
	if err != nil {
		h.l.Error("failed to generate tiny url", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, nil)
	}
	return ctx.JSON(http.StatusOK, &v0.GenerateURLResponse{
		ExpireTime:       tinyURL.ExpireTime.String(),
		GeneratedTinyURL: tinyURL.TinyURL,
	})
}

func NewHandler(logger *zap.Logger, s types.URLService) (types.Handler, error) {
	swagger, err := v0.GetSwagger()
	if err != nil {
		logger.Error("failed to get swagger", zap.Error(err))
		return nil, err
	}
	schema, err := types.LoadSchema(swagger)
	if err != nil {
		logger.Error("failed to load schema from openapi spec", zap.Error(err))
		return nil, err
	}
	return &handler{
		apiVersion: "v0",
		l:          logger,
		schema:     schema,
		svc:        s,
	}, nil
}
