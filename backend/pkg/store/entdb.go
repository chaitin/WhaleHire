package store

import (
	dql "database/sql"
	"log/slog"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/ptonlix/whalehire/backend/config"
	"github.com/ptonlix/whalehire/backend/db"
	_ "github.com/ptonlix/whalehire/backend/db/runtime"
)

func NewEntDB(cfg *config.Config, logger *slog.Logger) (*db.Client, error) {
	w, err := sql.Open(dialect.Postgres, cfg.Database.Master)
	if err != nil {
		return nil, err
	}
	w.DB().SetMaxOpenConns(cfg.Database.MaxOpenConns)
	w.DB().SetMaxIdleConns(cfg.Database.MaxIdleConns)
	w.DB().SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)
	r, err := sql.Open(dialect.Postgres, cfg.Database.Slave)
	if err != nil {
		return nil, err
	}

	r.DB().SetMaxOpenConns(cfg.Database.MaxOpenConns)
	r.DB().SetMaxIdleConns(cfg.Database.MaxIdleConns)
	r.DB().SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)
	c := db.NewClient(db.Driver(NewMultiDriver(r, w, logger)))
	if cfg.Debug {
		c = c.Debug()
	}

	return c, nil
}

func RecoverMigrate16(m *migrate.Migrate, logger *slog.Logger) {
	logger = logger.With("fn", "RecoverMigrate16")
	logger.Info("recover migrate 16")
	version, dirty, err := m.Version()
	if err != nil {
		logger.With("err", err).Error("get version failed")
		return
	}

	logger.With("version", version, "dirty", dirty).Info("get schema_migrations")
	if version == 16 && dirty {
		if err := m.Force(15); err != nil {
			logger.With("err", err).Error("force migrate 15 failed")
			return
		}
	}

	logger.Info("recover migrate 16 success")
}

func MigrateSQL(cfg *config.Config, logger *slog.Logger) error {
	db, err := dql.Open("postgres", cfg.Database.Master)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migration",
		"postgres", driver)
	if err != nil {
		return err
	}
	RecoverMigrate16(m, logger)
	if err := m.Up(); err != nil {
		logger.With("component", "db").With("err", err).Warn("migrate db failed")
	}

	return nil
}
