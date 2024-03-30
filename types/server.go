package types

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/labstack/echo/v4"

	"github.com/vaishakdinesh/tiny-url-svc/types/api/rest/v0"
)

type (
	MiddlewareFunc = echo.MiddlewareFunc

	// Registerer defines a interface to be able to register handlers to a server
	Registerer interface {
		Register(*Server)
	}

	// Handler represents the V0 API
	Handler interface {
		Registerer
		v0.ServerInterface
	}

	// OpenAPISchema is an abstraction for OAS 3.0 spec
	OpenAPISchema interface {
		ValidateRequest(req *http.Request) error
		ValidationMiddleware() MiddlewareFunc
	}

	// Server represents an HTTP server
	Server struct {
		*echo.Echo
	}

	// APIError defines the struct the handlers return
	APIError struct {
		Code    int
		Message string
	}

	openAPISchema3 struct {
		router routers.Router
	}
)

var (
	emptySpecError    = errors.New("empty oas spec")
	IntServerAPIError = &APIError{
		Code:    InternalServerError,
		Message: ErrInternalServer.Error(),
	}
)

func (ae *APIError) Error() string {
	return ae.Message
}

// LoadSchema loads the schema defined by the API spec to create a router based on the paths defined
func LoadSchema(oas *openapi3.T) (OpenAPISchema, error) {
	if oas == nil {
		return nil, emptySpecError
	}
	router, err := legacy.NewRouter(oas)
	if err != nil {
		return nil, err
	}
	return &openAPISchema3{
		router: router,
	}, nil
}

// ValidateRequest validates incoming http request
func (o *openAPISchema3) ValidateRequest(req *http.Request) error {
	route, pathParam, err := o.router.FindRoute(req)
	if err != nil {
		return err
	}
	validationRequest := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParam,
		Route:      route,
	}
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	return openapi3filter.ValidateRequest(ctx, validationRequest)
}

// ValidationMiddleware is the middle the server uses for API requests
func (o *openAPISchema3) ValidationMiddleware() MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := o.ValidateRequest(c.Request())
			if err != nil {
				return c.JSON(http.StatusBadRequest, &APIError{
					Code:    ValidationError,
					Message: err.Error(),
				})
			}
			err = next(c)
			if err != nil {
				if errors.Is(err, echo.ErrNotFound) {
					return c.JSON(http.StatusNotFound, &APIError{
						Code:    NoRouteError,
						Message: ErrNoPath.Error(),
					})
				}
			}
			return nil
		}
	}
}

// Run starts and runs the server
func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	go func() {
		<-ctx.Done()
		if err := s.Shutdown(ctx); err != nil {
			s.Logger.Errorf("failed to shut down server %s", err)
		}
	}()
	err := s.Start(":8000")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.Logger.Fatalf("server failed %s", err)
	}
}
