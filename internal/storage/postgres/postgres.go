package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexMickh/speak-auth/internal/config"
	"github.com/AlexMickh/speak-auth/internal/domain/models"
	"github.com/AlexMickh/speak-auth/internal/storage"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/jmoiron/sqlx"
)

type Postgres struct {
	db *sqlx.DB
}

func New(cfg config.DBConfig) (*Postgres, error) {
	const op = "storage.postgres.New"

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := sqlx.Connect("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db.SetMaxOpenConns(cfg.MaxPools)

	m, err := migrate.New(
		"file://"+cfg.MigrationsPath,
		connString,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) SaveUser(ctx context.Context, username, email, password string) (string, error) {
	const op = "storage.postgres.SaveUser"

	var id string
	err := p.db.GetContext(
		ctx, &id,
		`INSERT INTO users
		(id, username, email, password)
		VALUES (gen_random_uuid(), $1, $2, $3)
		RETURNING id`,
		username, email, password,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (p *Postgres) GetUser(ctx context.Context, email, password string) (*models.User, error) {
	const op = "storage.postgres.GetUser"

	var user models.User
	err := p.db.GetContext(
		ctx, &user,
		"SELECT * FROM users WHERE email = $1 AND password = $2",
		email, password,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (p *Postgres) Close() {
	p.db.Close()
}
