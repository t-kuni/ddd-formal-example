-- Local test data only. Example:
-- docker compose exec -T postgres psql -U orbit -d orbit -v customer_id="<UUID>" < scripts/expire-cart.sql
BEGIN;
UPDATE aggregates
SET body=jsonb_set(body,'{expires}',to_jsonb(extract(epoch from now())::bigint - 1)),version=version+1
WHERE kind='cart' AND id=:'customer_id'::uuid;
INSERT INTO projection_events(id,kind,aggregate_id,version,body)
SELECT gen_random_uuid(),kind,id,version,body FROM aggregates WHERE kind='cart' AND id=:'customer_id'::uuid;
COMMIT;
