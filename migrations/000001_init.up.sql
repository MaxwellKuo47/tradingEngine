Table users {
  id bigserial [pk]
  created_at timestamp [not null, default: `now()`]
  name text [not null]
  email text [unique, not null]
  password_hash bytea [not null]
  version integer [not null, default: 1]
}

Table tokens {
  hash bytea [pk]
  user_id bigint [not null, ref: > users.id]
  expiry timestamp [not null]
}

Table orders {
  id bigserial [pk]
  user_id bigint [not null, ref: > users.id]
  stock_id bigint [not null, ref: > stocks.id]
  type integer [not null, note:"0: buy 1: sell"]
  quantity integer [not null]
  price_type integer [not null, note:"0: market 1: limit"]
  price decimal [null, note:"null for market orders"]
  status integer [not null, note:"-1: killed 0: pending 1: filled"] // 
  created_at timestamp [not null, default: `now()`]
  updated_at timestamp [not null, default: `now()`]
  version integer [not null, default: 1]
}

Table trades {
  id bigserial [pk]
  buy_order_id bigint [not null, ref: > orders.id]
  sell_order_id bigint [not null, ref: > orders.id]
  quantity integer [not null]
  price decimal [not null]
  executed_at timestamp [not null, default: `now()`]
}

Table stocks {
  id bigserial [pk]
  destination string [not null]
}