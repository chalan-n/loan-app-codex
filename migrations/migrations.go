package migrations

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"loan-app/models"

	"gorm.io/gorm"
)

type Migration struct {
	Version string
	Name    string
	Up      func(tx *gorm.DB) error
}

func list() []Migration {
	return []Migration{
		{
			Version: "2026040801",
			Name:    "core_schema",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&models.User{},
					&models.LoanApplication{},
					&models.Guarantor{},
					&models.AuditLog{},
					&models.WebAuthnCredential{},
					&models.RefRunning{},
				)
			},
		},
		{
			Version: "2026040802",
			Name:    "loan_file_metadata",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.LoanFile{})
			},
		},
	}
}

func validate(migrations []Migration) error {
	seen := map[string]struct{}{}
	prev := ""
	for _, migration := range migrations {
		if migration.Version == "" || migration.Name == "" || migration.Up == nil {
			return errors.New("invalid migration definition")
		}
		if _, ok := seen[migration.Version]; ok {
			return fmt.Errorf("duplicate migration version: %s", migration.Version)
		}
		if prev != "" && migration.Version < prev {
			return fmt.Errorf("migrations out of order: %s before %s", migration.Version, prev)
		}
		seen[migration.Version] = struct{}{}
		prev = migration.Version
	}
	return nil
}

func Run(db *gorm.DB) error {
	if db == nil {
		return errors.New("database is nil")
	}

	migrations := list()
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	if err := validate(migrations); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.SchemaMigration{}); err != nil {
		return err
	}

	applied := map[string]struct{}{}
	var rows []models.SchemaMigration
	if err := db.Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		applied[row.Version] = struct{}{}
	}

	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; ok {
			continue
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := migration.Up(tx); err != nil {
				return err
			}
			return tx.Create(&models.SchemaMigration{
				Version:   migration.Version,
				Name:      migration.Name,
				AppliedAt: time.Now().UTC(),
			}).Error
		}); err != nil {
			return fmt.Errorf("apply migration %s_%s: %w", migration.Version, migration.Name, err)
		}
	}

	return nil
}
