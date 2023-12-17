CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "name" text NOT NULL,
  "email" text UNIQUE NOT NULL,
  "password_hash" bytea NOT NULL,
  "version" integer NOT NULL DEFAULT 1
);

CREATE TABLE "tokens" (
  "hash" bytea PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "expiry" timestamp NOT NULL
);

CREATE TABLE "orders" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "stock_id" bigint NOT NULL,
  "type" integer NOT NULL,
  "quantity" integer NOT NULL,
  "price_type" integer NOT NULL,
  "price" decimal,
  "status" integer NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  "version" integer NOT NULL DEFAULT 1
);

CREATE TABLE "trades" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "order_id" bigint NOT NULL,
  "quantity" integer NOT NULL,
  "price" decimal NOT NULL,
  "executed_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "user_stock_blances" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint UNIQUE NOT NULL,
  "stock_id" bigint NOT NULL,
  "quantity" integer NOT NULL,
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  "version" integer NOT NULL DEFAULT 1
);

CREATE TABLE "user_wallets" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "blance" decimal NOT NULL,
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  "version" integer NOT NULL DEFAULT 1
);

CREATE TABLE "stocks" (
  "id" bigserial PRIMARY KEY,
  "name" text NOT NULL
);

CREATE INDEX ON "orders" ("user_id");

CREATE INDEX ON "orders" ("stock_id");

CREATE INDEX ON "orders" ("user_id", "stock_id");

CREATE INDEX ON "orders" ("user_id", "status");

CREATE INDEX ON "trades" ("user_id");

CREATE INDEX ON "trades" ("order_id");

CREATE INDEX ON "trades" ("user_id", "order_id");

CREATE INDEX ON "user_stock_blances" ("user_id");

CREATE INDEX ON "user_stock_blances" ("stock_id");

CREATE INDEX ON "user_stock_blances" ("user_id", "stock_id");

CREATE INDEX ON "user_wallets" ("user_id");

COMMENT ON COLUMN "orders"."type" IS '0: buy 1: sell';

COMMENT ON COLUMN "orders"."price_type" IS '0: market 1: limit';

COMMENT ON COLUMN "orders"."price" IS 'null for market orders';

COMMENT ON COLUMN "orders"."status" IS '-1: killed 0: pending 1: filled';

ALTER TABLE "tokens" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "orders" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "orders" ADD FOREIGN KEY ("stock_id") REFERENCES "stocks" ("id");

ALTER TABLE "trades" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "trades" ADD FOREIGN KEY ("order_id") REFERENCES "orders" ("id");

ALTER TABLE "user_stock_blances" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "user_stock_blances" ADD FOREIGN KEY ("stock_id") REFERENCES "stocks" ("id");

ALTER TABLE "user_wallets" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
