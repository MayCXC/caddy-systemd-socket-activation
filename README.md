# caddy-systemd-socket-activation
`sd` and `sdgram` custom networks for caddy

a container image can be built with `xcaddy` and then tagged like so:

```sh
podman build -f - -t caddy-sdsa . <<-'EOT'
	FROM docker.io/caddy:2-builder AS builder
	RUN xcaddy build master --with github.com/MayCXC/caddy-systemd-socket-activation
	FROM docker.io/caddy:2
	COPY --from=builder /usr/bin/caddy /usr/bin/caddy
	COPY Caddyfile /etc/caddy/Caddyfile
	EOT
```
