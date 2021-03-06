# README #

`quicklog` is a recent event store that allows fast browsing of application event traces.

A sample front-end using `quicklog` is the [github.com/karmakaze/quickvue](https://github.com/karmakaze/quickvue) project.

```
time       |sources...                      |details...
published  |web-browser api-server image-svc|type (action)       actor     object      target context (extra info)
-----------|----------- ---------- ---------|------------------- --------- ----------- ------ ------------------------------
12:18.123   1.2.3.4                          click:upload-button user:1234                    user-agent=...
12:18.234               api-22               POST /uploads       user:1234                    filename=cat.jpg filesize=9876
12:18.567                          img-01    POST /images        user:1234 img:45512          size=9876
```

### What can I do with it? ###

* Collect events from all sources: client (web/mobile), edge, internal, async tasks, batch jobs
* See chrnological sequence diagram for a filtered subset of events
* (coming soon) get recent metrics/statistics
* build front-ends for interacting with collected data

### Schema

Below is the schema for the main `entry` table. See [schema.sql](schema.sql) for full details.

```
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
```

### Entry Examples

| source              | actor     | type   | object        | target    | context              |
| ------------------- | --------  | ------ | ------------- | --------- | -------------------- |
| ip:100.101.102.103  | user:1234 | click  | button:upload | null      | {"page": "/photos"}  |
| host:api.myapp.site | user:1234 | upload | file:logo.png | null      | null                 |
| host:imgserver.site | user:1234 | create | image:123     | null      | {"file": "logo.png"} |

(note: this doesn't quite match the sequence diagram shown at top of README)

### Building ###

* [Download and install the Go compiler version 1.9 or later](https://golang.org/dl/)
* `make build`
* or `make build-linux # for cross-compilation

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

### Development ###

* `go get github.com/codegangsta/gin`
* `PATH` has to include `$GOPATH/bin` # usually /home/&lt;username&gt;/go/bin
* `gin -p 8124 -a 3000 run main.go`

The above will rebuild/restart `quicklog` whenever the .go source changes.
The `gin` watcher listens on port 8124 (normally the quicklog port) as a proxy and forwards requests on port 3000.
Quicklog will listen on port 3000 as `gin` will set `PORT` accordingly.

### Smoke Test ###

* `curl -s 'http://localhost:8124/entries' |./jl`
* `curl -si -X POST -H 'content-type: application/json' `
   `-d '{"project_id": 1, "published": "2018-08-13T02:13:12.713221Z", "source": "a-source", `
   `"type": "an-action", "actor": "an-actor", `
   `"object": "an-object", "target": "a-target", "context": {"string": "value", "number": 1, `
   `"boolean": true, "null": null, "object": {"list": []}, "list": [{}], "Pi": 3.14159}, `
   `"trace_id": "a-trace-id", "span_id": "a-span-id"}' 'http://localhost:8124/entries'`
* `curl -s 'http://localhost:8124/entries' |./jl`

### How to run tests ###

* coming soon...

### Contribution guidelines ###

* Writing tests
* Code review
* Other guidelines

### Who do I talk to? ###

* Repo owner or admin
* Other community or team contact


This software is Copyright (c) 2018, Keith Kim. All rights reserved.
License information can be found in the LICENSE file.
