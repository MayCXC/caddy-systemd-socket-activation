# caddy-systemd-socket-activation
`sd` and `sdgram` custom networks for caddy

a container image can be built with `xcaddy`:

```sh
nerdctl build -t caddy-sdsa -f - . <<-EOT
	FROM docker.io/caddy:2-builder AS builder
	RUN xcaddy build master --with github.com/MayCXC/caddy-systemd-socket-activation
	FROM caddy:2
	COPY --from=builder /usr/bin/caddy /usr/bin/caddy
	EOT
```
