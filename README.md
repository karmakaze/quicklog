# README #

This README would normally document whatever steps are necessary to get your application up and running.

### What is this repository for? ###

* Quick summary
* Version
* [Learn Markdown](https://bitbucket.org/tutorials/markdowndemo)

### Schema

Below is the schema for the main `entry` table. See [schema.sql](schema.sql) for full details.

```
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
```

### Entry Examples

| source              | actor     | type   | object        | target_id | context              |
| ------------------- | --------  | ------ | ------------- | --------- | -------------------- |
| ip:100.101.102.103  | user:1234 | click  | button:upload | null      | {"page": "/photos"}  |
| host:api.myapp.site | user:1234 | upload | file:logo.png | null      | null                 |
| host:imgserver.site | user:1234 | create | image:123     | null      | {"file": "logo.png"} |

### How do I get set up? ###

* Summary of set up
* Configuration
* Dependencies
* Database configuration
* How to run tests
* Deployment instructions

### Contribution guidelines ###

* Writing tests
* Code review
* Other guidelines

### Who do I talk to? ###

* Repo owner or admin
* Other community or team contact
