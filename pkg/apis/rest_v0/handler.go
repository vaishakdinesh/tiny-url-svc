package rest_v0

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/vaishakdinesh/tiny-url-svc/types"
	v0 "github.com/vaishakdinesh/tiny-url-svc/types/api/rest/v0"
)

const cacheExpire = time.Hour * 24 * 7

type handler struct {
	apiVersion  string
	schema      types.OpenAPISchema
	svc         types.URLService
	l           *zap.Logger
	redisClient *redis.Client
}

func NewHandler(logger *zap.Logger, s types.URLService, r *redis.Client) (types.Handler, error) {
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
		apiVersion:  "v0",
		l:           logger,
		schema:      schema,
		svc:         s,
		redisClient: r,
	}, nil
}

func (h *handler) Register(s *types.Server) {
	sg := s.Group("/api/" + h.apiVersion)
	sg.Use(h.schema.ValidationMiddleware())
	v0.RegisterHandlers(sg, h)
}

func (h *handler) GenerateURL(ctx echo.Context) error {
	genURLReq, err := decodeRequest(ctx)
	if err != nil {
		h.l.Error("failed to decode request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, err)
	}

	reqCTX := ctx.Request().Context()
	cached, err := h.checkCache(reqCTX, genURLReq.Url)
	if err != nil {
		h.l.Error("failed to get cache", zap.Error(err))
	}
	if cached != nil {
		return ctx.JSON(http.StatusOK, cached)
	}
	h.l.Info(fmt.Sprintf("url %s not in cache", genURLReq.Url))

	tinyURL, err := h.svc.GenerateTinyURL(reqCTX, genURLReq.Url)
	if err != nil {
		h.l.Error("failed to generate tiny url", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, &types.APIError{
			Code:    types.InternalServerError,
			Message: err.Error(),
		})
	}

	err = h.Add(reqCTX, genURLReq.Url, tinyURL)
	if err != nil {
		h.l.Error(fmt.Sprintf("failed to add %s to cache", genURLReq.Url), zap.Error(err))
	}
	h.l.Info(fmt.Sprintf("url %s cached", genURLReq.Url))

	return ctx.JSON(http.StatusOK, &v0.GenerateURLResponse{
		ExpireTime:       tinyURL.ExpireTime.String(),
		GeneratedTinyURL: tinyURL.TinyURL,
	})
}

// TODO:: invalidate cache if expired
func (h *handler) checkCache(ctx context.Context, key string) (*v0.GenerateURLResponse, error) {
	cached, err := h.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	cachedURL := new(types.URLDocument)
	err = json.Unmarshal([]byte(cached), cachedURL)
	if err != nil {
		return nil, err
	}
	return &v0.GenerateURLResponse{
		ExpireTime:       cachedURL.ExpireTime.String(),
		GeneratedTinyURL: cachedURL.TinyURL,
	}, nil
}

func (h *handler) Add(ctx context.Context, key string, val types.URLDocument) error {
	keyBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	statusCMD := h.redisClient.Set(ctx, key, keyBytes, cacheExpire)
	if statusCMD.Err() != nil {
		return err
	}
	return nil
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
