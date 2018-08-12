CREATE TABLE entry (
  id        bigserial NOT NULL,
  published timestamp NOT NULL,
  source    varchar   NOT NULL,
  type      varchar   NOT NULL,
  actor     varchar   NOT NULL,
  object    varchar   NOT NULL,
  target    varchar   NOT NULL,
  context   jsonb,
  trace_id  varchar   NOT NULL,
  span_id   varchar   NOT NULL
);
