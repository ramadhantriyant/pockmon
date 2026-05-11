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
			account.DELETE("/:id", h.DeactivateAccount)
			account.GET("/:id/adjustment", h.ListAdjustmentsByAccount)
			account.POST("/:id/adjustment", h.CreateAdjustment)
		}

		recurring := api.Group("/recurring")
		{
			recurring.GET("", h.ListRecurringTransactions)
			recurring.GET("/active", h.ListActiveRecurringTransactions)
			recurring.GET("/:id", h.GetRecurringTransaction)
			recurring.POST("", h.CreateRecurringTransaction)
			recurring.PUT("/:id", h.UpdateRecurringTransaction)
			recurring.DELETE("/:id", h.DeactivateRecurringTransaction)
		}

		adjustment := api.Group("/adjustment")
		{
			adjustment.GET("", h.ListAdjustments)
			adjustment.GET("/:id", h.GetAdjustment)
		}

		notification := api.Group("/notification")
		{
			notification.GET("", h.ListNotifications)
			notification.GET("/unread", h.ListUnreadNotifications)
			notification.PATCH("/:id/read", h.MarkNotificationAsRead)
			notification.PATCH("/read-all", h.MarkAllNotificationsAsRead)
			notification.DELETE("/:id", h.DeleteNotification)
			notification.DELETE("/read", h.DeleteReadNotifications)
		}

		goal := api.Group("/goal")
		{
			goal.GET("", h.ListGoals)
			goal.GET("/active", h.ListActiveGoals)
			goal.GET("/type/:type", h.ListGoalsByType)
			goal.GET("/:id", h.GetGoal)
			goal.GET("/:id/progress", h.GetGoalProgress)
			goal.POST("", h.CreateGoal)
			goal.PUT("/:id", h.UpdateGoal)
			goal.PATCH("/:id/contribute", h.ContributeToGoal)
			goal.PATCH("/:id/complete", h.CompleteGoal)
			goal.DELETE("/:id", h.DeleteGoal)
		}

		budget := api.Group("/budget")
		{
			budget.GET("", h.ListBudgets)
			budget.GET("/active", h.ListActiveBudgets)
			budget.GET("/spending", h.ListBudgetsWithSpending)
			budget.GET("/alerts", h.ListBudgetsExceedingThreshold)
			budget.GET("/:id", h.GetBudget)
			budget.GET("/:id/spending", h.GetBudgetWithSpending)
			budget.POST("", h.CreateBudget)
			budget.PUT("/:id", h.UpdateBudget)
			budget.DELETE("/:id", h.DeactivateBudget)
		}

		transfer := api.Group("/transfer")
		{
			transfer.GET("", h.ListTransfers)
			transfer.GET("/:id", h.GetTransfer)
			transfer.GET("/account/:id", h.ListTransfersByAccount)
			transfer.POST("", h.CreateTransfer)
			transfer.DELETE("/:id", h.DeleteTransfer)
		}

		transaction := api.Group("/transaction")
		{
			transaction.GET("", h.ListTransactions)
			transaction.GET("/category/:id", h.ListTransactionsByCategory)
			transaction.GET("/account/:id", h.ListTransactionsByAccount)
			transaction.GET("/type/:type", h.ListTransactionsByType)
			transaction.GET("/tags", h.ListTransactionsByTags)
			transaction.GET("/:id", h.GetTransaction)
			transaction.POST("", h.CreateTransaction)
			transaction.PUT("/:id", h.UpdateTransaction)
			transaction.DELETE("/:id", h.DeleteTransaction)
			transaction.GET("/:id/attachment", h.ListAttachmentsByTransaction)
			transaction.GET("/:id/attachment/upload-url", h.GetUploadURL)
			transaction.POST("/:id/attachment", h.ConfirmAttachment)
		}

		attachment := api.Group("/attachment")
		{
			attachment.GET("/:id", h.GetAttachment)
			attachment.DELETE("/:id", h.DeleteAttachment)
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
