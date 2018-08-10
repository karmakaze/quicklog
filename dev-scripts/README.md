# github-webhook-server.py

This script will rebuild the head of the `master` branch and restart the `quicklog` service.

To use it:

- `mkdir /opt/quicklog`
- `cd /opt/quicklog`
- `git clone git@github.com:karmakaze/quicklog .`
- mkdir -p ~/go/src/github.com/karmakaze
- ln -s /opt/quicklog ~/go/src/github.com/karmakaze/quicklog
- `make build`
- install the `quicklog.service` to work with `systemd` changing the `user` value as necessary.

Then still from the `/opt/quicklog` directory, run:

- `nohup dev-scripts/github-webhook-server.py &`
- Add a github webhook to your server (default is HTTP port 9854)

The `dev-scripts/start-webhook-server.sh` script will do just this (also stopping any previously running instance).
