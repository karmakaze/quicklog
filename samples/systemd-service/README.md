# quicklog.service - README.md

Installing `quicklog` as a service for `systemd`:

- make build
- sudo mkdir -p /opt/quicklog
- sudo cp quicklog /opt/quicklog/
- sudo chown -R ubuntu:ubuntu /opt/quicklog
- sudo cp quicklog.service /etc/systemd/system/
- sudo chown root:root /etc/systemd/system/quicklog.service
- sudo systemctl daemon-reload
- sudo systemctl enable quicklog
- sudo service quicklog start
- sudo tail -f /var/log/syslog
