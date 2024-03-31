package rest_v0

import (
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vaishakdinesh/tiny-url-svc/types"
	v0 "github.com/vaishakdinesh/tiny-url-svc/types/api/rest/v0"
)

type handler struct {
	schema types.OpenAPISchema
	svc    types.URLService
	l      *zap.Logger
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
		l:      logger,
		schema: schema,
		svc:    s,
	}, nil
}

func (h *handler) Register(s *types.Server) {
	sg := s.Group("/tinyurlsvc")
	sg.Use(h.schema.ValidationMiddleware())
	v0.RegisterHandlers(sg, h)
}

func (h *handler) GenerateURL(ctx echo.Context) error {
	genURLReq, err := decodeRequest(ctx)
	if err != nil {
		h.l.Error("failed to decode request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, &types.APIError{
			Code:    types.InputError,
			Message: err.Error(),
		})
	}

	tinyURL, err := h.svc.GenerateTinyURL(ctx.Request().Context(), genURLReq.Url)
	if err != nil {
		h.l.Error("failed to generate tiny url", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, &types.APIError{
			Code:    types.InternalServerError,
			Message: err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, &v0.GenerateURLResponse{
		ExpireTime:       tinyURL.ExpireTime.String(),
		GeneratedTinyURL: tinyURL.ToURL(ctx),
	})
}

func (h *handler) GetURL(ctx echo.Context, urlKey string) error {
	urlDoc, err := h.svc.GetTinyURL(ctx.Request().Context(), urlKey)
	if err != nil {
		h.l.Error("failed to get tiny url from db", zap.Error(err))
		return ctx.JSON(http.StatusNotFound, &types.APIError{
			Code:    types.NotFoundError,
			Message: err.Error(),
		})
	}
	http.Redirect(ctx.Response().Unwrap(), ctx.Request(), urlDoc.LongURL, http.StatusMovedPermanently)
	return nil
}

func (h *handler) DeleteURL(ctx echo.Context, urlKey string) error {
	err := h.svc.DeleteTinyURL(ctx.Request().Context(), urlKey)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrCacheNotFound):
			return ctx.JSON(http.StatusNoContent, nil)
		case errors.Is(err, types.ErrDocumentNotFound):
			return ctx.JSON(http.StatusNotFound, &types.APIError{
				Code:    types.NotFoundError,
				Message: err.Error(),
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, &types.APIError{
				Code:    types.InternalServerError,
				Message: err.Error(),
			})
		}
	}
	return ctx.JSON(http.StatusNoContent, nil)
}

func decodeRequest(ctx echo.Context) (*v0.GenerateURLRequest, error) {
	genURLReq := new(v0.GenerateURLRequest)
	err := json.NewDecoder(ctx.Request().Body).Decode(genURLReq)
	if err != nil {
		return nil, &types.APIError{
			Code:    types.InputError,
			Message: err.Error(),
		}
	}
	return genURLReq, nil
}
