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

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/db"
	_ "github.com/chaitin/WhaleHire/backend/db/runtime"
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

// ProblematicMigration 定义有问题的迁移版本及其恢复策略
type ProblematicMigration struct {
	Version        int    // 有问题的版本号
	ForceToVersion int    // 强制回退到的版本号
	Description    string // 问题描述
}

// getProblematicMigrations 返回已知有问题的迁移版本列表
func getProblematicMigrations() []ProblematicMigration {
	return []ProblematicMigration{
		// 未来如果发现其他有问题的迁移版本，可以在这里添加
		// {
		//     Version:        XX,
		//     ForceToVersion: XX-1,
		//     Description:    "description of the issue",
		// },
	}
}

// RecoverMigration 通用的迁移恢复函数，处理已知有问题的迁移版本
func RecoverMigration(m *migrate.Migrate, logger *slog.Logger) {
	logger = logger.With("fn", "RecoverMigration")
	logger.Info("starting migration recovery check")

	version, dirty, err := m.Version()
	if err != nil {
		logger.With("err", err).Error("get version failed")
		return
	}

	logger.With("version", version, "dirty", dirty).Info("current migration status")

	// 如果迁移状态不是dirty，则无需恢复
	if !dirty {
		logger.Info("migration is clean, no recovery needed")
		return
	}

	// 检查当前版本是否在已知问题列表中
	problematicMigrations := getProblematicMigrations()
	for _, pm := range problematicMigrations {
		if int(version) == pm.Version {
			logger.With("version", version, "description", pm.Description).
				Warn("found problematic migration, attempting recovery")

			if err := m.Force(pm.ForceToVersion); err != nil {
				logger.With("err", err, "force_to_version", pm.ForceToVersion).
					Error("force migration failed")
				return
			}

			logger.With("version", version, "forced_to", pm.ForceToVersion).
				Info("migration recovery successful")
			return
		}
	}

	// 如果当前版本不在已知问题列表中，记录警告但不进行强制恢复
	logger.With("version", version).
		Warn("migration is dirty but not in known problematic list, manual intervention may be required")
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
	RecoverMigration(m, logger)
	if err := m.Up(); err != nil {
		logger.With("component", "db").With("err", err).Warn("migrate db failed")
	}

	return nil
}
