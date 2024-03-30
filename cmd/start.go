package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/vaishakdinesh/tiny-url-svc/pkg/apis/rest_v0"
	"github.com/vaishakdinesh/tiny-url-svc/pkg/db"
	"github.com/vaishakdinesh/tiny-url-svc/pkg/url"
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

	// TODO:: from config
	opt, err := redis.ParseURL("redis://tsvc:tsvcPassword@redis:6379/0")
	if err != nil {
		logger.Fatal("failed parse redis url", zap.Error(err))
	}

	redisClient := redis.NewClient(opt)
	defer func() {
		if err = redisClient.Close(); err != nil {
			logger.Fatal("failed to close redis cache", zap.Error(err))
		}
	}()

	server, err := newServer()
	if err != nil {
		logger.Fatal("failed to create a new server", zap.Error(err))
	}

	apis, err := initHandlers(logger, dbClient, redisClient)
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

func initHandlers(l *zap.Logger, c *mongo.Client, r *redis.Client) ([]types.Registerer, error) {
	urlSvc := url.NewTinyURLService(l, db.NewURLRepo(c))
	tinyURLV0, err := rest_v0.NewHandler(l, urlSvc, r)
	if err != nil {
		return nil, err
	}
	return []types.Registerer{tinyURLV0}, nil
}

func initDatastore(ctx context.Context, logger *zap.Logger) (*mongo.Client, error) {
	return db.NewDB(ctx, logger)
}
