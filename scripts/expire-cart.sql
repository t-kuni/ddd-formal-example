-- 検証環境専用。実行後にmaintenanceを起動する。
-- docker compose exec -T postgres psql -U orbit -d orbit -v customer_id="<顧客UUID>" < scripts/expire-cart.sql
UPDATE carts
SET expires_at = floor(extract(epoch from now()))::bigint - 1,
    version = version + 1
WHERE customer_id = :'customer_id'::uuid;
