package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhantriyant/pockmon/internal/config"
	"github.com/ramadhantriyant/pockmon/internal/handler"
	"github.com/ramadhantriyant/pockmon/internal/middleware"
)

func createServer(ctx context.Context, cfg *config.Config, port string) *http.Server {
	authClient, err := cfg.FirebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("failed to create Firebase auth client: %v", err)
	}
	cfg.AuthClient = authClient

	r := gin.New()
	r.Use(
		gin.Recovery(),
		middleware.Logger(),
		middleware.ErrorHandler(),
	)
	h := handler.New(cfg)

	auth := r.Group("/auth", middleware.Auth(authClient))
	{
		auth.POST("/register", h.Register)
		auth.GET("/me", h.GetMe)
		auth.PUT("/me/currency/:code", h.UpdateCurrency)
	}

	user := r.Group("/user", middleware.Auth(authClient))
	{
		user.GET("", h.ListUsers)
		user.PUT("/:id", h.ToggleAdmin)
		user.DELETE("/:id", h.DeleteUser)
	}

	api := r.Group("/api", middleware.Auth(authClient))
	{
		category := api.Group("/category")
		{
			category.GET("", h.ListCategories)
			category.GET("/:id", h.GetCategory)
			category.POST("", h.CreateCategory)
			category.PUT("/:id", h.UpdateCategory)
			category.DELETE("/:id", h.DeleteCategory)
		}

		account := api.Group("/account")
		{
			account.GET("", h.ListAccounts)
			account.GET("/:id", h.GetAccount)
			account.POST("", h.CreateAccount)
			account.PUT("/:id", h.UpdateAccount)
		}

		transaction := api.Group("/transaction")
		{
			transaction.GET("", h.ListTransactions)
			transaction.GET("/category/:id", h.ListTransactionsByCategory) // List by category
			transaction.GET("/account/:id", h.ListTransactionsByAccount)   // List by account
			transaction.GET("/type/:type", h.ListTransactionsByType)       // List by type (expense, income)
			transaction.GET("/tags", h.ListTransactionsByTags)             // List by tag
			transaction.POST("", h.CreateTransaction)
			transaction.PUT("/:id", h.UpdateTransaction)
			transaction.DELETE("/:id", h.DeleteTransaction)
		}
	}
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":    404,
			"error":     "not found",
			"message":   "the requested route does not exist",
			"path":      c.Request.URL.Path,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	return &http.Server{
		Addr:    port,
		Handler: r,
	}
}

func runServer(ctx context.Context, server *http.Server, shutdownTimeout time.Duration) error {
	serverErr := make(chan error, 1)

	go func() {
		log.Println("starting server...")
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case <-stop:
		log.Println("shutting down...")
	case <-ctx.Done():
		log.Println("context canceled")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		if closeErr := server.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}

	log.Println("shutdown completed")
	return nil
}
