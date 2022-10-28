# Listen for a valid TOTP token and exit

This small program will wait for a valid TOTP token to be sent on a TCP port
and then will exit. This is useful for exposing single-user services on
Internet.

## Service lifecycle (systemd)

- Original service startup is delayed by a call to [waitotp], this is achieved
  by overriding parts of the unit file (see [waitotp.conf] for example)
- [waitotp] listens on original service port but does not expose the service
  itself and does not reply to any incoming packets
- If a valid TOTP token is received, [waitotp] exits and lets the original
  service to proceed with startup
- Immediately after original service becomes active, systemd launches the
  related [inactivity-@.service]
- [inactivity.sh] tracks network usage of the original service and if
  original service is inactive for a specified amount of time,
  [inactivity-@.service] restarts
- Systemd triggers a restart of original service which again ends up asking
  for a TOTP token before proceeding

[waitotp]: waitotp.go
[waitotp.conf]: waitotp.conf
[inactivity-@.service]: inactivity-@.service
[inactivity.sh]: inactivity.sh
