package postgresql

import (
	"context"
	"errors"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/lib"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

// ErrorClassification тип для классификации ошибок
type PGErrorClassification int

const (
	// NonRetriable - операцию не следует повторять
	NonRetriable PGErrorClassification = iota

	// Retriable - операцию можно повторить
	Retriable
)

// PostgresErrorClassifier классификатор ошибок PostgreSQL
type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify классифицирует ошибку и возвращает PGErrorClassification
func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}

	// Проверяем и конвертируем в pgconn.PgError, если это возможно
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return СlassifyPgError(pgErr)
	}

	// По умолчанию считаем ошибку неповторяемой
	return NonRetriable
}

func СlassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	// Коды ошибок PostgreSQL: https://www.postgresql.org/docs/current/errcodes-appendix.html

	switch pgErr.Code {
	// Класс 08 - Ошибки соединения
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable

	// Класс 40 - Откат транзакции
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return Retriable

	// Класс 57 - Ошибка оператора
	case pgerrcode.CannotConnectNow: // 57P03
		return Retriable
	}

	// Можно добавить более конкретные проверки с использованием констант pgerrcode
	switch pgErr.Code {
	// Класс 22 - Ошибки данных
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetriable

	// Класс 23 - Нарушение ограничений целостности
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetriable

	// Класс 42 - Синтаксические ошибки
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetriable
	}

	// По умолчанию считаем ошибку неповторяемой
	return NonRetriable
}

var pgToDomain = map[string]error{
	"23505": domain.ErrLoginAlreadyTaken, // UniqueViolation
	"22023": domain.ErrInvalidPayload,    // InvalidParameterValue
	"P0002": domain.ErrUserNotFound,      // NoDataFound
}

func sqlState(err error) (string, bool) {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) {
		return pgxErr.Code, true
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code), true
	}
	return "", false
}

func translate(op string, err error) error {
	if err == nil {
		return nil
	}
	if code, ok := sqlState(err); ok {
		if srv, ok2 := pgToDomain[code]; ok2 {
			return domain.MakeError(lib.StandardError(op, err), srv)
		}
		return domain.MakeError(lib.StandardError(op, err), domain.ErrInternal)
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return domain.MakeError(lib.StandardError(op, err), domain.ErrServiceUnavailable)
	}
	return domain.MakeError(lib.StandardError(op, err), domain.ErrInternal)
}
