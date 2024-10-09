# caddy-systemd-socket-activation
a plugin that adds `sd` and `sdgram` custom networks for caddy.

an example Caddyfile:

```
{
	auto_https disable_redirects
	admin off
}

http://localhost {
	bind sd/caddy.socket/0 {
		protocols h1
	}
	log
	respond "Hello, HTTP!"
}

https://localhost {
	bind sd/caddy.socket/1 {
		protocols h1 h2
	}
	bind sdgram/CaddyDatagram/0 {
		protocols h3
	}
	log
	respond "Hello, HTTPS!"
}
```

can be used with `xcaddy` from its working directory to build and tag a container image that uses this plugin, like so:

```sh
podman build -f - -t caddy-sdsa . <<-'EOT'
	FROM docker.io/caddy:2-builder AS builder
	RUN xcaddy build master --with github.com/MayCXC/caddy-systemd-socket-activation
	FROM docker.io/caddy:2
	COPY --from=builder /usr/bin/caddy /usr/bin/caddy
	COPY Caddyfile /etc/caddy/Caddyfile
	EOT
```

then systemd socket and service units can be used to activate a container created from it:

`caddy.service`:

```
[Unit]
Description=Caddy
Documentation=https://caddyserver.com/docs/
After=network.target network-online.target
Requires=network-online.target

[Service]
Type=notify
User=caddy
Group=caddy
Environment=PODMAN_SYSTEMD_UNIT=%n
Restart=on-failure
ExecStart=podman run --rm localhost/caddy-sdsa
TimeoutStopSec=5s
LimitNOFILE=1048576
PrivateTmp=true
ProtectSystem=full
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
```

`caddy.socket`:

```
[Socket]
ListenStream=80
ListenStream=443

[Install]
WantedBy = sockets.target
```

`caddyh3.socket`:

```
[Socket]
ListenDatagram=443
Service=caddy.service
FileDescriptorName=CaddyDatagram

[Install]
WantedBy = sockets.target
```

the modified `caddy` binary can also tested from the systemd host via a bind mount with:

`systemd-socket-activate -l 80 -l 443 systemd-socket-activate -l 443 -d -E LISTEN_FDNAMES="caddy.socket:caddy.socket:CaddyDatagram" ./caddy run`

and podman >=4.0.0 can take advantage of quadlets to make configuration less hectic, see https://github.com/eriksjolund/podman-caddy-socket-activation .
