# caddy-systemd-socket-activation
a plugin that adds `sd` and `sdgram` custom networks for caddy.

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

from a working directory containing this Caddyfile,`xcaddy` can be used to a container image can be build and tag a container image like so:

```sh
podman build -f - -t caddy-sdsa . <<-'EOT'
	FROM docker.io/caddy:2-builder AS builder
	RUN xcaddy build master --with github.com/MayCXC/caddy-systemd-socket-activation
	FROM docker.io/caddy:2
	COPY --from=builder /usr/bin/caddy /usr/bin/caddy
	COPY Caddyfile /etc/caddy/Caddyfile
	EOT
```
