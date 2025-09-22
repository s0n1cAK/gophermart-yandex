-- Удаляем триггер (иначе нельзя дропнуть функцию)
DROP TRIGGER IF EXISTS apply_balance_entry_after_ins ON user_balance_entries;

DROP PROCEDURE IF EXISTS public.proc_register_user;
DROP PROCEDURE IF EXISTS public.update_user_pwd;
DROP FUNCTION IF EXISTS public.sfn_hash_password(text);
DROP FUNCTION IF EXISTS public.sfn_validate_user(varchar, text);
DROP FUNCTION IF EXISTS public.trg_apply_balance_entry();

DROP INDEX IF EXISTS idx_user_orders_user_time;
DROP INDEX IF EXISTS idx_user_orders_user_status;
DROP INDEX IF EXISTS idx_user_withdrawals_user_time;
DROP INDEX IF EXISTS idx_user_balance_entries_user_time;

DROP TABLE IF EXISTS user_withdrawals CASCADE;
DROP TABLE IF EXISTS user_point_balances CASCADE;
DROP TABLE IF EXISTS user_balance_entries CASCADE;
DROP TABLE IF EXISTS user_orders CASCADE;
DROP TABLE IF EXISTS users_credentials CASCADE;
DROP TABLE IF EXISTS users CASCADE;
