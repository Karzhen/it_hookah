package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/karzhen/restaurant-lk/internal/app"
	"github.com/karzhen/restaurant-lk/internal/auth"
	"github.com/karzhen/restaurant-lk/internal/config"
	"github.com/karzhen/restaurant-lk/internal/handler"
	"github.com/karzhen/restaurant-lk/internal/middleware"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		logger.Error("failed to connect db", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := dbPool.Ping(pingCtx); err != nil {
		logger.Error("failed to ping db", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(dbPool)
	refreshRepo := repository.NewRefreshTokenRepository(dbPool)
	catalogRepo := repository.NewCatalogRepository(dbPool)
	cartRepo := repository.NewCartRepository(dbPool)
	orderRepo := repository.NewOrderRepository(dbPool)
	mixRepo := repository.NewMixRepository(dbPool)
	tagRepo := repository.NewTagRepository(dbPool)
	stockMovementRepo := repository.NewStockMovementRepository(dbPool)

	passwordManager := auth.NewBcryptPasswordManager(bcrypt.DefaultCost)
	jwtManager := auth.NewJWTManager(cfg.JWT.AccessSecret, time.Duration(cfg.JWT.AccessTTLMinutes)*time.Minute)
	refreshManager := auth.NewRefreshTokenManager(time.Duration(cfg.JWT.RefreshTTLHours)*time.Hour, cfg.JWT.RefreshSecret)

	authService := service.NewAuthService(userRepo, refreshRepo, passwordManager, jwtManager, refreshManager, logger)
	userService := service.NewUserService(userRepo)
	catalogService := service.NewCatalogService(catalogRepo)
	cartService := service.NewCartManager(cartRepo, catalogRepo)
	orderService := service.NewOrderManager(orderRepo)
	mixService := service.NewMixManager(mixRepo, catalogRepo)
	tagService := service.NewTagManager(tagRepo, catalogRepo, mixRepo)
	stockMovementService := service.NewStockMovementManager(stockMovementRepo, catalogRepo)

	authHandler := handler.NewAuthHandler(authService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	healthHandler := handler.NewHealthHandler()
	cartHandler := handler.NewCartHandler(cartService, catalogRepo, logger)
	orderHandler := handler.NewOrderHandler(orderService, userService, logger)
	mixHandler := handler.NewMixHandler(mixService, catalogRepo, logger)
	tagHandler := handler.NewTagHandler(tagService, logger)
	stockMovementHandler := handler.NewStockMovementHandler(stockMovementService, logger)
	catalogPublicHandler := handler.NewCatalogPublicHandler(catalogService, logger)
	catalogAdminHandler := handler.NewCatalogAdminHandler(catalogService, logger)

	authMW := middleware.NewAuthMiddleware(jwtManager, userService, logger)
	router := app.NewRouter(app.RouterDependencies{
		AuthHandler:         authHandler,
		UserHandler:         userHandler,
		HealthHandler:       healthHandler,
		CartHandler:         cartHandler,
		OrderHandler:        orderHandler,
		MixHandler:          mixHandler,
		TagHandler:          tagHandler,
		StockMovementHandle: stockMovementHandler,
		CatalogPublicHandle: catalogPublicHandler,
		CatalogAdminHandle:  catalogAdminHandler,
		AuthMW:              authMW,
		Logger:              logger,
	})

	httpServer := app.NewHTTPServer(cfg, router, logger)

	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.Start(); err != nil {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", "error", err)
		}
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout())
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
