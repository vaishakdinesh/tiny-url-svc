package cmd

import (
	"context"
	"github.com/vaishakdinesh/tiny-url-svc/pkg/url"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"

	"github.com/vaishakdinesh/tiny-url-svc/pkg/apis/rest_v0"
	"github.com/vaishakdinesh/tiny-url-svc/pkg/db"
	"github.com/vaishakdinesh/tiny-url-svc/types"
)

type health struct {
	Status string `json:"status"`
}

// Run initializes all the dependencies and starts a go routine for the server
func Run() {
	ctx := context.Background()
	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}

	dbClient, err := initDatastore(ctx, logger)
	if err != nil {
		logger.Fatal("failed to create db", zap.Error(err))
	}
	logger.Info("connected to DB")
	defer func() {
		if err = dbClient.Disconnect(ctx); err != nil {
			logger.Fatal("failed to disconnect db", zap.Error(err))
		}
	}()

	server, err := newServer()
	if err != nil {
		logger.Fatal("failed to create a new server", zap.Error(err))
	}

	apis, err := initHandlers(logger, dbClient)
	if err != nil {
		logger.Fatal("failed to init rest handlers", zap.Error(err))
	}
	for _, a := range apis {
		a.Register(server)
	}

	wg := new(sync.WaitGroup)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	wg.Add(1)
	server.Run(ctx, wg)
	wg.Wait()
}

func newServer() (*types.Server, error) {
	s := &types.Server{
		Echo: echo.New(),
	}
	s.GET("/healthy", func(c echo.Context) error {
		return c.JSON(http.StatusOK, &health{Status: "healthy"})
	})
	return s, nil
}

func initHandlers(l *zap.Logger, c *mongo.Client) ([]types.Registerer, error) {
	urlSvc := url.NewTinyURLService(l, db.NewURLRepo(c))
	tinyURLV0, err := rest_v0.NewHandler(l, urlSvc)
	if err != nil {
		return nil, err
	}
	return []types.Registerer{tinyURLV0}, nil
}

func initDatastore(ctx context.Context, logger *zap.Logger) (*mongo.Client, error) {
	return db.NewDB(ctx, logger)
}
