CREATE TABLE project (
  id      serial PRIMARY KEY,
  name    varchar NOT NULL
);

CREATE UNIQUE INDEX project_name_idx ON project (name);

CREATE TABLE entry (
  project_id     integer     NOT NULL,
  seq            bigserial   NOT NULL,
  published      timestamptz NOT NULL,
  source         varchar     NOT NULL,
  type           varchar     NOT NULL,
  actor          varchar     NOT NULL,
  object         varchar     NOT NULL,
  target         varchar     NOT NULL,
  context        jsonb,
  trace_id       varchar,
  parent_span_id varchar,
  span_id        varchar,

  PRIMARY KEY (project_id, seq)
);

CREATE INDEX entry_object_idx ON entry (object);
CREATE INDEX entry_target_idx ON entry (target);
CREATE INDEX entry_trace_id_idx ON entry (trace_id) WHERE trace_id IS NOT NULL;
CREATE INDEX entry_parent_span_id_idx ON entry (parent_span_id) WHERE parent_span_id IS NOT NULL;
CREATE INDEX entry_span_id_idx ON entry (span_id) WHERE span_id IS NOT NULL;

CREATE TABLE span_tag (
  project_id integer NOT NULL,
  trace_id   varchar NOT NULL,
  span_id    varchar NOT NULL,
  key        varchar NOT NULL,
  value      varchar NOT NULL,

  PRIMARY KEY (project_id, value, key, span_id)
);
