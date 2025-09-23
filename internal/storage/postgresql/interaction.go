package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/models"
)

func (s *PostgresStorage) CheckUser(ctx context.Context, login string) (bool, error) {
	var exist bool

	err := retryWrapper(ctx, func() error {
		return s.Database.QueryRowContext(ctx,
			`SELECT
				CASE WHEN EXISTS 
				(
					SELECT * FROM public.users  WHERE login_name=$1
				)
				THEN 'TRUE'
				ELSE 'FALSE'
			END`,
			login,
		).Scan(&exist)
	})
	if err != nil {
		return false, translate("postgresql.GetUserByLogin.select", err)
	}

	return exist, nil
}

func (s *PostgresStorage) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	var user models.User

	err := retryWrapper(ctx, func() error {
		return s.Database.QueryRowContext(ctx,
			`SELECT id, login_name FROM users WHERE login_name = $1`,
			login,
		).Scan(&user.ID, &user.Login)
	})
	if err != nil {
		return models.User{}, translate("postgresql.GetUserByLogin.select", err)
	}

	return user, nil
}

func (s *PostgresStorage) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	var user models.User

	err := retryWrapper(ctx, func() error {
		return s.Database.QueryRowContext(ctx,
			`SELECT id, login_name FROM users WHERE id = $1`,
			id,
		).Scan(&user.ID, &user.Login)
	})
	if err != nil {
		return models.User{}, translate("postgresql.GetUserByLogin.select", err)
	}

	return user, nil
}

func (s *PostgresStorage) RegisterUser(ctx context.Context, u models.User) error {
	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			return err
		}

		defer func() { _ = tx.Rollback() }()

		_, err = tx.ExecContext(ctx, `
			WITH inserted_user AS (
				INSERT INTO users (login_name)
				VALUES ($1)
				RETURNING id
			)
			INSERT INTO users_credentials (user_id, password_hash)
			SELECT id, sfn_hash_password($2)
			FROM inserted_user;
		`, u.Login, u.Password)

		if err != nil {
			return err
		}
		return tx.Commit()
	})
	if err != nil {
		return translate("postgresql.RegisterUser", err)
	}
	return nil

}

func (s *PostgresStorage) ValidateUser(ctx context.Context, u models.User) error {
	var ok bool

	err := retryWrapper(ctx, func() error {
		return s.Database.QueryRowContext(ctx,
			`
			SELECT COALESCE(
				crypt($2, c.password_hash) = c.password_hash,
				false
			) AS is_valid
			FROM users u
			JOIN users_credentials c ON c.user_id = u.id
			WHERE u.login_name = trim($1)
			LIMIT 1;
			`,
			u.Login, u.Password,
		).Scan(&ok)
	})
	if err != nil {
		return translate("postgresql.ValidateUser.select", err)
	}

	if !ok {
		return domain.MakeError(fmt.Errorf("postgresql.ValidateUser doesn't match"), domain.ErrInvalidCredentials)

	}

	return nil
}

func (s *PostgresStorage) CheckOrder(ctx context.Context, user string, order models.Order) error {
	var exist bool

	err := retryWrapper(ctx, func() error {
		return s.Database.QueryRowContext(ctx, `
			SELECT
				CASE 
					WHEN EXISTS (SELECT 1 FROM public.user_orders WHERE order_number = $1)
						THEN 'TRUE'
				ELSE 'FALSE'                              
			END;`,
			order.Number,
		).Scan(&exist)
	})
	if err != nil {
		return translate("postgresql.CheckOrder.select", err)
	}

	if exist {
		var output string

		err := retryWrapper(ctx, func() error {
			return s.Database.QueryRowContext(ctx, `
				SELECT 
					CASE
						WHEN EXISTS (SELECT 1 FROM public.user_orders WHERE order_number = $1 AND user_id = (SELECT id FROM users WHERE login_name = $2))
							THEN 'already_yours'
						WHEN EXISTS (SELECT * FROM public.user_orders WHERE order_number = $1 AND user_id <> (SELECT id FROM users WHERE login_name = $2))
							THEN 'belongs_to_other'
					ELSE 'unexpected'
				END;`,
				order.Number, user,
			).Scan(&output)
		})
		if err != nil {
			return translate("postgresql.CheckOrder.select", err)
		}

		if output == "already_yours" {
			return domain.MakeError(fmt.Errorf("postgresql.CheckOrder order already created by user"), domain.ErrOrderCreatedByUser)
		}
		if output == "belongs_to_other" {
			return domain.MakeError(fmt.Errorf("postgresql.CheckOrder order already created by other user"), domain.ErrOrderCreatedByOtherUser)
		}
		if output == "unexpected" {
			return domain.MakeError(fmt.Errorf("postgresql.CheckOrder something strange happens"), domain.ErrInternal)
		}
	}

	return nil
}

func (s *PostgresStorage) CreateOrder(ctx context.Context, user string, order models.Order) error {
	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		result, err := tx.ExecContext(ctx,
			`WITH u AS (
			 	SELECT id FROM users WHERE login_name = $1
			 )
			 INSERT INTO user_orders (user_id, order_number)
			 SELECT u.id, $2 FROM u
			 RETURNING id;`,
			user, order.Number,
		)

		if err != nil {
			return err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return domain.MakeError(fmt.Errorf("postgresql.CreateOrder Check Rows"), domain.ErrInternal)
		}
		if affected == 0 {
			return domain.MakeError(fmt.Errorf("postgresql.CreateOrder Check User"), domain.ErrUserNotFound)
		}
		return tx.Commit()
	})

	if err != nil {
		return translate("postgresql.CreateOrder", err)
	}
	return nil
}

func (s *PostgresStorage) FetchNewOrders(ctx context.Context, limit int) ([]models.Order, error) {
	orders := make([]models.Order, 0, limit)

	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		rows, err := tx.QueryContext(ctx,
			`WITH ts AS (SELECT now() AS ts)
			 UPDATE user_orders u
			 SET status = 'PROCESSING',
			 	processing_started_at = ts.ts
			 FROM ts
			 WHERE u.id IN (
			 	SELECT id
			 	FROM user_orders
			 	WHERE status = 'NEW'
			 	ORDER BY created_at ASC
			 	FOR UPDATE SKIP LOCKED
			 	LIMIT $1
			 )
			 RETURNING u.order_number, u.user_id, u.points_awarded, u.created_at;`,
			limit,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var o models.Order
			err := rows.Scan(&o.Number, &o.UserID, &o.Accrual, &o.UploadedAt)
			if err != nil {
				return err
			}
			o.Status = "PROCESSING"
			orders = append(orders, o)
		}

		err = rows.Err()
		if err != nil {
			return err
		}

		return tx.Commit()
	})

	if err != nil {
		return []models.Order{}, translate("postgresql.FetchNewOrders", err)
	}
	return orders, nil
}

func (s *PostgresStorage) FetchProccesingOrders(ctx context.Context, limit int) ([]models.Order, error) {
	orders := make([]models.Order, 0, limit)

	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		rows, err := tx.QueryContext(ctx,
			`WITH ts AS (SELECT now() AS ts)
			 UPDATE user_orders u
			 SET processing_started_at = ts.ts
			 FROM ts
			 WHERE u.id IN (
			 	SELECT id
			 	FROM user_orders
			 	WHERE status = 'PROCESSING'
			 	AND (
			 		processing_started_at IS NULL OR 
					processing_started_at < ts.ts - interval '5 minutes'
			 	)
			 	ORDER BY
					processing_started_at NULLS FIRST,
					created_at
			 	FOR UPDATE SKIP LOCKED
			 	LIMIT $1
			 )
			 RETURNING u.order_number, u.user_id, u.points_awarded, u.created_at;`,
			limit,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var o models.Order
			err := rows.Scan(&o.Number, &o.UserID, &o.Accrual, &o.UploadedAt)
			if err != nil {
				return err
			}

			orders = append(orders, o)
		}

		err = rows.Err()
		if err != nil {
			return err
		}

		return tx.Commit()
	})

	if err != nil {
		return []models.Order{}, translate("postgresql.FetchProccesingOrders", err)
	}
	return orders, nil
}

func (s *PostgresStorage) UpdateOrderProcessed(ctx context.Context, orderNumber string, points float64) error {
	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		result, err := tx.ExecContext(ctx, `
		UPDATE user_orders
		SET status = 'PROCESSED',
		    points_awarded = $2
		WHERE order_number = $1 AND status = 'PROCESSING'`, orderNumber, points)
		if err != nil {
			return err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return domain.MakeError(fmt.Errorf("postgresql.CreateOrder Check Rows"), domain.ErrInternal)
		}
		if affected == 0 {
			return nil
		}

		return tx.Commit()
	})
	if err != nil {
		return translate("postgresql.UpdateOrderProcessed", err)
	}
	return nil
}

func (s *PostgresStorage) UpdateOrderInvalid(ctx context.Context, orderNumber string) error {
	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		result, err := tx.ExecContext(ctx, `
		UPDATE user_orders
		SET status = 'INVALID'
		WHERE order_number = $1 AND status = 'PROCESSING'`, orderNumber)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return domain.MakeError(fmt.Errorf("postgresql.CreateOrder Check Rows"), domain.ErrInternal)
		}
		if affected == 0 {
			return tx.Rollback()
		}

		return tx.Commit()
	})
	if err != nil {
		return translate("postgresql.UpdateOrderInvalid", err)
	}
	return nil
}

func (s *PostgresStorage) GetOrders(ctx context.Context, user string) ([]models.Order, error) {
	orders := make([]models.Order, 0)

	err := retryWrapper(ctx, func() error {
		rows, err := s.Database.QueryContext(ctx,
			`
			SELECT order_number, status, points_awarded, created_at 
			FROM public.user_orders 
			WHERE user_id = (SELECT id FROM public.users WHERE login_name = $1)
			ORDER BY created_at DESC`,
			user,
		)

		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var o models.Order
			err = rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt)
			if err != nil {
				return err
			}

			orders = append(orders, o)
		}

		err = rows.Err()
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.Order{}, translate("postgresql.GetOrders", err)
	}

	return orders, nil
}

func (s *PostgresStorage) UpdateBalanceEntries(ctx context.Context, order models.Order) error {
	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		result, err := tx.ExecContext(ctx, `
		INSERT INTO user_balance_entries (user_id, entry_type, amount_points, order_id)
		SELECT uo.user_id, 'accrual', uo.points_awarded, uo.id
		FROM user_orders uo
		WHERE uo.status = 'PROCESSED'
		  AND uo.user_id = $1
		  AND uo.order_number = $2
		ON CONFLICT (order_id, entry_type) DO NOTHING;
		`, order.UserID, order.Number)
		if err != nil {
			return err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return domain.MakeError(fmt.Errorf("postgresql.CreateOrder Check Rows"), domain.ErrInternal)
		}
		if affected == 0 {
			return tx.Rollback()
		}

		return tx.Commit()
	})
	if err != nil {
		return translate("postgresql.UpdateBalanceEntries", err)
	}

	if err = s.UpdateBalance(ctx, order.UserID); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) UpdateMissingBalanceEntries(ctx context.Context) error {
	var users []uint64

	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		rows, err := tx.QueryContext(ctx, `
		WITH inserted AS (
			INSERT INTO user_balance_entries (user_id, entry_type, amount_points, order_id)
			SELECT uo.user_id, 'accrual', uo.points_awarded, uo.id
			FROM user_orders uo
			LEFT JOIN user_balance_entries ube
			ON ube.order_id = uo.id AND ube.entry_type = 'accrual'
			WHERE uo.status = 'PROCESSED'
			  AND ube.order_id IS NULL
			ON CONFLICT (order_id, entry_type) DO NOTHING
			RETURNING user_id
		)
		SELECT DISTINCT user_id FROM inserted;
		`)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var user uint64
			err := rows.Scan(&user)
			if err != nil {
				return err
			}
			users = append(users, user)
		}

		err = rows.Err()
		if err != nil {
			return err
		}

		return tx.Commit()
	})

	if err != nil {
		return translate("postgresql.UpdateMissingBalanceEntries", err)
	}

	for _, u := range users {
		if err = s.UpdateBalance(ctx, u); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) UpdateBalance(ctx context.Context, user uint64) error {
	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		result, err := tx.ExecContext(ctx, `
		INSERT INTO user_point_balances (user_id, balance, withdrawal, updated_at)
		SELECT 
			upb.user_id,
			COALESCE(SUM(CASE WHEN upb.entry_type = 'accrual' THEN upb.amount_points ELSE 0 END), 0)
			- COALESCE(SUM(CASE WHEN upb.entry_type = 'withdrawal' THEN upb.amount_points ELSE 0 END), 0) AS balance,
			COALESCE(SUM(CASE WHEN upb.entry_type = 'withdrawal' THEN upb.amount_points ELSE 0 END), 0) AS withdrawal,
			now()
		FROM user_balance_entries AS upb
		WHERE upb.user_id = $1
		GROUP BY upb.user_id
		ON CONFLICT (user_id) DO UPDATE
		SET balance   = EXCLUDED.balance,
			withdrawal = EXCLUDED.withdrawal,
			updated_at = EXCLUDED.updated_at;
		`, user)
		if err != nil {
			return err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return domain.MakeError(fmt.Errorf("postgresql.CreateOrder Check Rows"), domain.ErrInternal)
		}
		if affected == 0 {
			return nil
		}

		return tx.Commit()
	})
	if err != nil {
		return translate("postgresql.UpdateBalance", err)
	}

	return nil
}

func (s *PostgresStorage) UpdateWithdrawlEntries(ctx context.Context, userID uint64, withdraw models.Withdrawal) error {
	err := retryWrapper(ctx, func() error {
		tx, err := s.Database.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		result, err := tx.ExecContext(ctx, `
		INSERT INTO user_balance_entries (user_id, entry_type, amount_points, withdrawal_ref)
		VALUES ($1, 'withdrawal', $2, $3)
		ON CONFLICT (order_id, entry_type) DO NOTHING;
		`, userID, withdraw.Sum, withdraw.Order)
		if err != nil {
			return err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return domain.MakeError(fmt.Errorf("postgresql.UpdateWithdrawlEntries Check Rows"), domain.ErrInternal)
		}
		if affected == 0 {
			return tx.Rollback()
		}

		return tx.Commit()
	})
	if err != nil {
		return translate("postgresql.UpdateWithdrawlEntries", err)
	}

	if err = s.UpdateBalance(ctx, userID); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) GetBalance(ctx context.Context, userID uint64) (models.Balance, error) {
	var balance models.Balance

	err := retryWrapper(ctx, func() error {
		return s.Database.QueryRowContext(ctx, `
			WITH ins AS (
				INSERT INTO user_point_balances (user_id, balance, withdrawal)
				VALUES ($1, 0, 0)
				ON CONFLICT (user_id) DO NOTHING
				RETURNING balance, withdrawal
			)
			SELECT balance, withdrawal
			FROM ins
			UNION ALL
			SELECT balance, withdrawal
			FROM user_point_balances
			WHERE user_id=$1
			LIMIT 1;
		`, userID).Scan(&balance.Current, &balance.Withdrawn)
	})

	if err != nil {
		return models.Balance{}, translate("postgresql.GetBalance", err)
	}

	return balance, nil
}

func (s *PostgresStorage) GetWithdrawls(ctx context.Context, user uint64) ([]models.Withdrawal, error) {
	withdrawals := make([]models.Withdrawal, 0)

	err := retryWrapper(ctx, func() error {
		rows, err := s.Database.QueryContext(ctx,
			`
			SELECT withdrawal_ref, amount_points, posted_at 
			FROM public.user_balance_entries
			WHERE user_id=$1 AND
			      entry_type='withdrawal'
			ORDER BY posted_at DESC`,
			user,
		)

		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var w models.Withdrawal
			err = rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt)
			if err != nil {
				return err
			}

			withdrawals = append(withdrawals, w)
		}

		err = rows.Err()
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.Withdrawal{}, translate("postgresql.GetWithdrawls", err)
	}

	return withdrawals, nil
}
