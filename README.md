# README #

This README would normally document whatever steps are necessary to get your application up and running.

### What is this repository for? ###

* Quick summary
* Version
* [Learn Markdown](https://bitbucket.org/tutorials/markdowndemo)

### Schema

```
create table entry (
  id        bigserial not null,
  at        timestamp with time zone not null,
  source_id varchar not null,
  actor_id  varchar not null,
  action    varchar not null,
  object_id varchar,
  target_id varchar,
  trace_id  varchar,
  span_id   varchar,
  context   varchar
);
```

### Entry Examples

| source_id           | actor_id  | action | object_id     | object_id     | target_id | context                             |
| ------------------- | --------  | ------ | ------------- | ------------- | --------- | ----------------------------------- |
| ip:100.101.102.103  | user:1234 | click  | button:upload | file:logo.png |   null    | {"page":"https://myapp.site/photos" |
| host:api.myapp.site | user:1234 | create | image:123     |   null        |   null    |   null                              |

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
