package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	appservices "github.com/zaidmasri/business-planning-tool/internal/application/services"
	"github.com/zaidmasri/business-planning-tool/internal/domain/ports"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/config"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/email"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
	bgworker "github.com/zaidmasri/business-planning-tool/internal/interface/worker"
)

func worker(dbPath string, interval time.Duration, cfg config.Config) {
	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer conn.Close()

	if err := sqlite.RunMigrations(conn); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	userRepo := sqlite.NewUserRepository(conn)
	planRepo := sqlite.NewPlanRepository(conn)
	inviteRepo := sqlite.NewInviteRepository(conn)
	outboxRepo := sqlite.NewOutboxRepository(conn)

	notifications := appservices.NewNotificationService(buildMailer(cfg), userRepo, planRepo, inviteRepo, cfg.AppBaseURL)
	w := bgworker.NewOutboxWorker(outboxRepo, notifications, interval)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("worker: polling outbox every %s", interval)
	w.Run(ctx)
	log.Println("worker: shutting down")
}

func buildMailer(cfg config.Config) ports.Mailer {
	if cfg.SMTPHost == "" {
		log.Println("worker: SMTP_HOST not set, using console mailer (emails will be logged, not sent)")
		return email.NewConsoleMailer()
	}
	return email.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)
}
