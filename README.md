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

| source              | actor     | type   | object        | target    | context              |
| ------------------- | --------  | ------ | ------------- | --------- | -------------------- |
| ip:100.101.102.103  | user:1234 | click  | button:upload | null      | {"page": "/photos"}  |
| host:api.myapp.site | user:1234 | upload | file:logo.png | null      | null                 |
| host:imgserver.site | user:1234 | create | image:123     | null      | {"file": "logo.png"} |

### Building ###

* make deps
* make build
* make build-linux # for cross-compilation

### Creating the database ###

* psql # as a superuser (default is `postgres`)
* \# sometimes: `sudo su postgres -c psql`
* CREATE DATABASE quicklog;
* CREATE USER quicklog WITH PASSWORD 'quicklog';
* GRANT ALL PRIVILEGES ON DATABASE quicklog TO quicklog;
* \q
* psql -h localhost -U quicklog "quicklog"
* -- copy/paste the contents of `schema.sql` at the above the `psql` prompt.
* \q

### Running ###

* ./quicklog &
* listens on tcp port 8124

To rebuild and restart:

* make build && ./restart.sh

### Deployment ###

* \# assuming you have `/etc/hosts` and `~/.ssh/config` set up for host `prod`
* ssh prod
* sudo vi /etc/systemd/system/quicklog.service
* \# copy/paste [quicklog.service](quicklog.service) into editor above, save and exit editor
* \# you can either use a different user than `quicklog` in the .service file
* \# or create the user `quicklog` on the system
* sudo mkdir -p /opt/quicklog
* sudo chown quicklog:quicklog /opt/quicklog
* sudo systemctl daemon-reload
* sudo systemctl enable quicklog # configure to autostart
* exit # logout from host `prod`

On your development workstation:

* make deploy # see the `deploy:` recipe in the [Makefile](Makefile)
* \# this will make a linux binary, install it as prod:/opt/quicklog/quicklog
* \# and start (or restart) the service

### How to run tests ###

* coming soon...

### Contribution guidelines ###

* Writing tests
* Code review
* Other guidelines

### Who do I talk to? ###

* Repo owner or admin
* Other community or team contact
