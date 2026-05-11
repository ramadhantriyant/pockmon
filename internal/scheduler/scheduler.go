package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"github.com/ramadhantriyant/pockmon/internal/database"
	"github.com/ramadhantriyant/pockmon/internal/util"
)

func nextDate(current time.Time, frequency string) time.Time {
	switch frequency {
	case "daily":
		return current.AddDate(0, 0, 1)
	case "weekly":
		return current.AddDate(0, 0, 7)
	case "biweekly":
		return current.AddDate(0, 0, 14)
	case "monthly":
		return current.AddDate(0, 1, 0)
	case "quarterly":
		return current.AddDate(0, 3, 0)
	case "yearly":
		return current.AddDate(1, 0, 0)
	default:
		return current.AddDate(0, 1, 0)
	}
}

func runDaily(ctx context.Context, db *pgxpool.Pool) {
	today := time.Now().UTC()
	processRecurring(ctx, db, today)
	generateNotifications(ctx, db, today)
}

// --- Recurring transaction processing ---

func processRecurring(ctx context.Context, db *pgxpool.Pool, today time.Time) {
	todayDate := pgtype.Date{Time: today, Valid: true}

	querier := database.New(db)
	due, err := querier.ListAutoCreateDueRecurringTransactions(ctx, todayDate)
	if err != nil {
		log.Printf("scheduler: failed to list due recurring transactions: %v", err)
		return
	}

	for _, r := range due {
		if err := processSingle(ctx, db, r, today); err != nil {
			log.Printf("scheduler: failed to process recurring transaction %v: %v", r.ID, err)
		}
	}

	log.Printf("scheduler: processed %d recurring transactions", len(due))
}

func processSingle(ctx context.Context, db *pgxpool.Pool, r database.ListAutoCreateDueRecurringTransactionsRow, today time.Time) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := database.New(tx)

	account, err := q.GetAccountByID(ctx, database.GetAccountByIDParams{
		ID:     r.AccountID,
		UserID: r.UserID,
	})
	if err != nil {
		return err
	}

	currentBal, err := account.CurrentBalance.Float64Value()
	if err != nil {
		return err
	}

	amt, err := r.Amount.Float64Value()
	if err != nil {
		return err
	}

	var newBalFloat float64
	if r.Type == "income" {
		newBalFloat = currentBal.Float64 + amt.Float64
	} else {
		newBalFloat = currentBal.Float64 - amt.Float64
	}

	newBalance, err := util.GetNumeric(newBalFloat)
	if err != nil {
		return err
	}

	txnID, err := util.GenerateUUID()
	if err != nil {
		return err
	}

	currencyCode := "IDR"
	if r.CurrencyCode != nil {
		currencyCode = *r.CurrencyCode
	}

	isRecurring := true
	todayDate := pgtype.Date{Time: today, Valid: true}

	if _, err = q.CreateTransaction(ctx, database.CreateTransactionParams{
		ID:                     txnID,
		UserID:                 r.UserID,
		AccountID:              r.AccountID,
		CategoryID:             r.CategoryID,
		Type:                   r.Type,
		Amount:                 r.Amount,
		CurrencyCode:           currencyCode,
		TransactionDate:        todayDate,
		Description:            r.Description,
		Tags:                   []string{},
		IsRecurring:            &isRecurring,
		RecurringTransactionID: r.ID,
	}); err != nil {
		return err
	}

	if _, err = q.UpdateAccountBalance(ctx, database.UpdateAccountBalanceParams{
		ID:             r.AccountID,
		CurrentBalance: newBalance,
	}); err != nil {
		return err
	}

	next := nextDate(r.NextDueDate.Time, r.Frequency)
	if _, err = q.UpdateNextDueDate(ctx, database.UpdateNextDueDateParams{
		ID:          r.ID,
		NextDueDate: pgtype.Date{Time: next, Valid: true},
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// --- Notification generation ---

func generateNotifications(ctx context.Context, db *pgxpool.Pool, today time.Time) {
	querier := database.New(db)
	created := 0

	created += notifyBudgetAlerts(ctx, querier, today)
	created += notifyBillReminders(ctx, querier, today)
	created += notifyGoalMilestones(ctx, querier)

	log.Printf("scheduler: created %d notifications", created)
}

func createNotification(ctx context.Context, q database.Querier, userID pgtype.UUID, notifType, title, message string, relatedID pgtype.UUID, relatedType string) error {
	id, err := util.GenerateUUID()
	if err != nil {
		return err
	}
	rt := relatedType
	_, err = q.CreateNotification(ctx, database.CreateNotificationParams{
		ID:          id,
		UserID:      userID,
		Type:        notifType,
		Title:       title,
		Message:     message,
		RelatedID:   relatedID,
		RelatedType: &rt,
	})
	return err
}

func notifyBudgetAlerts(ctx context.Context, q database.Querier, today time.Time) int {
	// Check spending against budgets for the current calendar month.
	firstOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	budgets, err := q.ListAllBudgetsExceedingThreshold(ctx, database.ListAllBudgetsExceedingThresholdParams{
		TransactionDate:   pgtype.Date{Time: firstOfMonth, Valid: true},
		TransactionDate_2: pgtype.Date{Time: lastOfMonth, Valid: true},
	})
	if err != nil {
		log.Printf("scheduler: failed to list budgets exceeding threshold: %v", err)
		return 0
	}

	count := 0
	for _, b := range budgets {
		title := fmt.Sprintf("Budget Alert: %s", b.Name)
		message := fmt.Sprintf("Your \"%s\" budget has reached %d%% of its limit.", b.Name, b.SpentPercentage)
		if err := createNotification(ctx, q, b.UserID, "budget_alert", title, message, b.ID, "budget"); err != nil {
			log.Printf("scheduler: failed to create budget_alert notification for budget %v: %v", b.ID, err)
			continue
		}
		count++
	}
	return count
}

func notifyBillReminders(ctx context.Context, q database.Querier, today time.Time) int {
	todayDate := pgtype.Date{Time: today, Valid: true}

	due, err := q.ListDueRecurringTransactions(ctx, todayDate)
	if err != nil {
		log.Printf("scheduler: failed to list due recurring transactions for reminders: %v", err)
		return 0
	}

	count := 0
	for _, r := range due {
		// Only notify for manual-entry recurring transactions (auto_create handled separately).
		if r.AutoCreate != nil && *r.AutoCreate {
			continue
		}
		title := fmt.Sprintf("Bill Reminder: %s", r.Description)
		message := fmt.Sprintf("Your recurring %s \"%s\" is due today. Remember to record it.", r.Type, r.Description)
		if err := createNotification(ctx, q, r.UserID, "bill_reminder", title, message, r.ID, "recurring_transaction"); err != nil {
			log.Printf("scheduler: failed to create bill_reminder notification for recurring %v: %v", r.ID, err)
			continue
		}
		count++
	}
	return count
}

func notifyGoalMilestones(ctx context.Context, q database.Querier) int {
	goals, err := q.ListGoalsReachedTarget(ctx)
	if err != nil {
		log.Printf("scheduler: failed to list goals reached target: %v", err)
		return 0
	}

	count := 0
	for _, g := range goals {
		title := fmt.Sprintf("Goal Reached: %s", g.Name)
		message := fmt.Sprintf("Congratulations! You've reached your goal \"%s\". Don't forget to mark it as complete.", g.Name)
		if err := createNotification(ctx, q, g.UserID, "goal_milestone", title, message, g.ID, "goal"); err != nil {
			log.Printf("scheduler: failed to create goal_milestone notification for goal %v: %v", g.ID, err)
			continue
		}
		count++
	}
	return count
}

// --- Scheduler entry point ---

func Start(ctx context.Context, db *pgxpool.Pool) *cron.Cron {
	c := cron.New()
	c.AddFunc("0 0 * * *", func() {
		runDaily(ctx, db)
	})
	c.Start()
	log.Println("scheduler: started, runs daily at midnight UTC")
	return c
}
