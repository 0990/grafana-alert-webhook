# grafana-alert-webhook

Grafana Alerting webhook receiver for forwarding Grafana 11+ alert notifications to push channels.

Current supported pusher:

- WeCom / 企业微信应用消息

The project is structured so more pushers can be added later through the `pushers` config list.

## Features

- Supports Grafana 11+ unified alerting webhook payloads.
- Converts Grafana alert payloads into concise text summaries.
- Sends messages to one or more enabled pushers.
- Keeps backward compatibility with the old flat WeCom config shape.
- Provides a GitHub Actions workflow for publishing Docker images to GHCR on tag push.

## Endpoints

### `POST /grafana_alert`

Receives Grafana webhook payloads and forwards the summarized alert message.

Example Grafana contact point URL:

```text
http://your-host:1111/grafana_alert
```

Optional query parameters:

```text
http://your-host:1111/grafana_alert?touser=USER_ID&totag=TAG_ID
```

If neither `touser` nor `totag` is provided, the pusher default target from `config.json` is used.

### `POST /send`

Manual message sending endpoint.

```text
http://your-host:1111/send?content=hello&touser=USER_ID
```

## Configuration

Default config path:

```text
config.json
```

Use another config file:

```bash
grafana-alert-webhook -c /path/to/config.json
```

Recommended config shape:

```json
{
  "listen": ":1111",
  "pushers": [
    {
      "name": "default-wecom",
      "type": "wecom",
      "enabled": true,
      "config": {
        "corpid": "wwxxxxxxxxxxxxxxxx",
        "corpsecret": "your-wecom-app-secret",
        "agentid": 1000002,
        "touserdefault": "UserID",
        "totagdefault": ""
      }
    }
  ]
}
```

WeCom config fields:

- `corpid`: 企业微信企业 ID.
- `corpsecret`: 企业微信应用 Secret.
- `agentid`: 企业微信应用 AgentId.
- `touserdefault`: Default user target when request URL does not include `touser` or `totag`.
- `totagdefault`: Default tag target when request URL does not include `touser` or `totag`.

Legacy flat config is still accepted:

```json
{
  "listen": ":1111",
  "corpid": "wwxxxxxxxxxxxxxxxx",
  "corpsecret": "your-wecom-app-secret",
  "agentid": 1000002,
  "touserdefault": "UserID"
}
```

Do not commit real production secrets to a public repository.

## Run Locally

```bash
go test ./...
go run . -c config.json
```

The service listens on the `listen` address from config, for example `:1111`.

## Run With Docker

Build locally:

```bash
docker build -t grafana-alert-webhook:local .
```

Run with a mounted config file:

```bash
docker run --rm -p 1111:1111 \
  -v /path/to/config.json:/app/config.json:ro \
  grafana-alert-webhook:local
```

## Publish Docker Image

The workflow at `.github/workflows/docker.yml` publishes images to GitHub Container Registry when a tag is pushed.

Create and push a tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Published image names:

```text
ghcr.io/<owner>/<repo>:v1.0.0
ghcr.io/<owner>/<repo>:latest
```

No Docker Hub secrets are required. The workflow uses GitHub's built-in `GITHUB_TOKEN` with `packages: write` permission.

## Adding More Pushers

New pusher types should follow the existing shape:

```json
{
  "name": "example",
  "type": "new-type",
  "enabled": true,
  "config": {}
}
```

Implementation entry points:

- Add a config struct for the new pusher.
- Implement the `Pusher` interface in `pusher.go`.
- Register the new `type` in `NewPushService`.
