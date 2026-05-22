package repository

import (
	"github.com/RomanKovalev007/organization_service/internal/apperr"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"errors"
)

func wrapDBError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return apperr.ErrAlreadyExists
		case "23503": // foreign_key_violation
			return apperr.ErrNotFound
		}
	}

	return err
}
