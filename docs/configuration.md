# Configuration

## File Locations

The OpsSquad node stores its configuration and logs in standard system locations, depending on how it was installed.

### System Installation (Root)
- **Binaries**: `/usr/local/lib/opssquad/`
- **Configuration**: `/etc/opssquad/node.yaml`
- **Logs**: `/var/log/opssquad/node.log`

### User Installation (Non-Root)
- **Binaries**: `~/.local/lib/opssquad/`
- **Configuration**: `~/.config/opssquad/node.yaml`
- **Logs**: `~/.local/log/opssquad/node.log`

## Configuration File (`node.yaml`)

The `node.yaml` file controls the node's behavior. It is automatically generated during installation but can be modified manually if needed.

### Structure

```yaml
app:
  node_id: "your-node-id"
  api_key: "your-api-key"

logging:
  level: "info"
  file: "/var/log/opssquad/node.log"
```

### Parameters

| Parameter | Description |
|-----------|-------------|
| `app.node_id` | The unique identifier for this node (provided by dashboard). |
| `app.api_key` | The authentication key for communicating with the platform. |
| `logging.level` | Log verbosity. Options: `debug`, `info`, `warn`, `error`. Default: `info`. |
| `logging.file` | Absolute path to the log file. |
