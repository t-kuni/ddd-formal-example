CREATE TABLE IF NOT EXISTS actors(id uuid PRIMARY KEY, body jsonb NOT NULL);
CREATE TABLE IF NOT EXISTS products(id uuid PRIMARY KEY, body jsonb NOT NULL, image bytea NOT NULL);
CREATE TABLE IF NOT EXISTS coupons(id uuid PRIMARY KEY,code text UNIQUE NOT NULL, percent integer NOT NULL CHECK(percent BETWEEN 0 AND 100));
CREATE TABLE IF NOT EXISTS aggregates(kind text NOT NULL,id uuid NOT NULL,version bigint NOT NULL CHECK(version>0),body jsonb NOT NULL,PRIMARY KEY(kind,id));
CREATE TABLE IF NOT EXISTS projections(kind text NOT NULL,id uuid NOT NULL,version bigint NOT NULL,body jsonb NOT NULL,PRIMARY KEY(kind,id));
CREATE TABLE IF NOT EXISTS projection_events(id uuid PRIMARY KEY,kind text NOT NULL,aggregate_id uuid NOT NULL,version bigint NOT NULL,body jsonb NOT NULL,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS commands(id uuid PRIMARY KEY,actor_id uuid NOT NULL,type text NOT NULL,payload jsonb NOT NULL,fingerprint text NOT NULL,status text NOT NULL DEFAULT 'queued',result jsonb,error text NOT NULL DEFAULT '',attempts integer NOT NULL DEFAULT 0,last_dispatched timestamptz,created_at timestamptz NOT NULL DEFAULT now(),finished_at timestamptz);
CREATE INDEX IF NOT EXISTS pending_commands ON commands(created_at) WHERE status='queued';
CREATE TABLE IF NOT EXISTS invoices(id uuid PRIMARY KEY,order_id uuid UNIQUE NOT NULL,customer_id uuid NOT NULL,pdf bytea NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),sent_at timestamptz,attempts integer NOT NULL DEFAULT 0,last_error text NOT NULL DEFAULT '',lease_until timestamptz);
