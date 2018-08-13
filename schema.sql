CREATE TABLE project (
  id   serial primary key,
  name varchar NOT NULL
);

CREATE TABLE entry (
  project_id integer     NOT NULL,
  seq        bigserial   NOT NULL,
  published  timestamptz NOT NULL,
  source     varchar     NOT NULL,
  type       varchar     NOT NULL,
  actor      varchar     NOT NULL,
  object     varchar     NOT NULL,
  target     varchar     NOT NULL,
  context    jsonb,
  trace_id   varchar     NOT NULL,
  span_id    varchar     NOT NULL,

  PRIMARY KEY(project_id, seq)
);

CREATE INDEX entry_project_id_published_idx ON entry (project_id, published);
