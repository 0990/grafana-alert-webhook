# grafana-alert-webhook

Grafana Alerting webhook receiver for forwarding Grafana 11+ alert notifications to push channels.

Current supported pushers:

- WeCom / 企业微信应用消息
- XiaoShan / 小闪机器人 webhook

## Features

- Supports Grafana 11+ unified alerting webhook payloads.
- Converts Grafana alert payloads into concise text summaries.
- Sends messages to one or more enabled pushers.
- Provides a GitHub Actions workflow for publishing Docker images to GHCR on tag push.

## Endpoints

### `POST /wecom/grafana_alert`

Receives Grafana webhook payloads and forwards only through enabled WeCom pushers.

```text
http://your-host:1111/wecom/grafana_alert?touser=USER_ID&totag=TAG_ID
```

### `POST /xiaoshan/grafana_alert`

Receives Grafana webhook payloads and forwards only through enabled XiaoShan pushers. `touser` is treated as comma-separated XiaoShan `atUserIds`.

```text
http://your-host:1111/xiaoshan/grafana_alert?touser=USER_ID&atmobiles=13800000000&atall=false
```

Type-specific manual endpoints are also available:

```text
http://your-host:1111/wecom/send?content=hello&touser=USER_ID
http://your-host:1111/xiaoshan/send?content=hello&touser=USER_ID&atmobiles=13800000000
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
  "wecom": {
    "enable": true,
    "corpid": "wwxxxxxxxxxxxxxxxx",
    "corpsecret": "your-wecom-app-secret",
    "agentid": 1000002,
    "touserdefault": "UserID",
    "totagdefault": ""
  },
  "xiaoshan": {
    "enable": true,
    "accesstoken": "your-xiaoshan-access-token",
    "secret": "your-xiaoshan-secret",
    "msgtype": "text",
    "atall": false,
    "atuserids": [],
    "atmobiles": []
  }
}
```

WeCom config fields:

- `enable`: Whether to enable WeCom pushing.
- `corpid`: 企业微信企业 ID.
- `corpsecret`: 企业微信应用 Secret.
- `agentid`: 企业微信应用 AgentId.
- `touserdefault`: Default user target when request URL does not include `touser` or `totag`.
- `totagdefault`: Default tag target when request URL does not include `touser` or `totag`.

XiaoShan config fields:

- `enable`: Whether to enable XiaoShan pushing.
- `url`: Optional webhook base URL. Defaults to `https://wapi.zhimagame.net:6543/robot/webhook/v2`.
- `accesstoken`: XiaoShan robot access token.
- `secret`: XiaoShan robot security secret used for HMAC-SHA256 signing.
- `msgtype`: `text` or `markdown`. Defaults to `text`.
- `atuserids`: Default XiaoShan user IDs to mention.
- `atmobiles`: Default mobile numbers to mention.
- `atall`: Whether to mention everyone by default.
- `maxtextlength`: Text message chunk size. Defaults to `2000`.

The XiaoShan config parser also accepts aliases such as `access_token`, `webhook_url`, `msg_type`, `at_userids`, and `at_mobiles`.

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
