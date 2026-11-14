# Deployment

How to run Herald for a real business, as it stands today. Read
[SECURITY-MODEL.md](SECURITY-MODEL.md) alongside this.

The examples here are starting points. Adapt them to your server and test them before
you rely on them.

## The one rule

**Never expose `/admin` to the internet.** The console and its two write routes
(`/admin/config`, `/admin/knowledge`) have no sign-in yet. Anyone who can reach them
can rewrite what your officer tells customers. Until sign-in is built:

- run Herald bound to `127.0.0.1`, and
- put a proxy in front that passes only the public routes.

## 1. Build one binary

```bash
make build            # produces ./bin/herald
```

Herald is a single file with no runtime dependencies. Copy it to the server, for
example to `/usr/local/bin/herald`.

## 2. Run it as a service, on localhost

Create a user and a folder for the configuration:

```bash
sudo useradd --system --home /var/lib/herald --shell /usr/sbin/nologin herald
sudo install -d -o herald -g herald -m 0750 /var/lib/herald
```

A systemd unit, `/etc/systemd/system/herald.service`:

```ini
[Unit]
Description=Herald front-office officer
After=network-online.target
Wants=network-online.target

[Service]
User=herald
Group=herald
WorkingDirectory=/var/lib/herald
ExecStart=/usr/local/bin/herald serve -addr 127.0.0.1:8080 -config /var/lib/herald/config.json
EnvironmentFile=-/etc/herald/env
Restart=on-failure
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/var/lib/herald

[Install]
WantedBy=multi-user.target
```

API keys go in `/etc/herald/env`, readable by root only (`chmod 600`), one per line:

```
GROQ_API_KEY=...
```

Then:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now herald
curl -s http://127.0.0.1:8080/healthz      # prints: ok
```

## 3. Put a proxy in front

The proxy gives you HTTPS and decides which routes the world can reach. The public
routes are:

| Route | Purpose |
|---|---|
| `POST /api/chat` | The chat API the widget calls |
| `GET /api/greeting` | The officer's opening line |
| `GET /widget.js` | The embeddable widget |
| `GET /widget` | A standalone chat page (optional) |
| `GET /healthz` | Health check (optional) |

Everything else, `/admin` above all, stays closed.

An example with [Caddy](https://caddyserver.com/), which obtains the certificate itself:

```
chat.example.com {
	@public path /api/chat /api/greeting /widget.js /widget /healthz
	handle @public {
		reverse_proxy 127.0.0.1:8080
	}
	handle {
		respond 404
	}
}
```

Herald builds the widget's address from the request's host and the
`X-Forwarded-Proto` header, so the widget calls back over HTTPS when the proxy sets
that header. Caddy does so by default.

The widget on your website is then:

```html
<script src="https://chat.example.com/widget.js" defer></script>
```

## 4. Reach the console safely

Use an SSH tunnel from your own computer:

```bash
ssh -L 8080:127.0.0.1:8080 you@your-server
```

and open `http://localhost:8080/admin` in your browser. The console is reachable only
through your SSH login.

## 5. Look after it

- **Back up** `/var/lib/herald/config.json`. It is the whole of your business's setup.
- **The knowledge index is in memory.** It is rebuilt from the configuration at every
  start. With a hosted embedder, each start sends the knowledge passages to that
  provider again.
- **Update** by replacing the binary and restarting the service. Read
  [CHANGELOG.md](../CHANGELOG.md) first.
- **Watch the handoffs.** A question that goes to a person is a gap in your knowledge.
  Add the missing fact and the officer answers it next time.

## What is not here yet

There is no rate limit, no sign-in for the console, and no log of conversations. If
your site is busy or a target, add rate limiting at the proxy. These are open areas
for contributors; see [OPEN-QUESTIONS.md](OPEN-QUESTIONS.md).
