# Configuration

## File Locations

The FixPanic agent stores its configuration and logs in standard system locations, depending on how it was installed.

### System Installation (Root)
- **Binaries**: `/usr/local/lib/fixpanic/`
- **Configuration**: `/etc/fixpanic/agent.yaml`
- **Logs**: `/var/log/fixpanic/agent.log`

### User Installation (Non-Root)
- **Binaries**: `~/.local/lib/fixpanic/`
- **Configuration**: `~/.config/fixpanic/agent.yaml`
- **Logs**: `~/.local/log/fixpanic/agent.log`

## Configuration File (`agent.yaml`)

The `agent.yaml` file controls the agent's behavior. It is automatically generated during installation but can be modified manually if needed.

### Structure

```yaml
app:
  agent_id: "your-agent-id"
  api_key: "your-api-key"

logging:
  level: "info"
  file: "/var/log/fixpanic/agent.log"
```

### Parameters

| Parameter | Description |
|-----------|-------------|
| `app.agent_id` | The unique identifier for this agent (provided by dashboard). |
| `app.api_key` | The authentication key for communicating with the platform. |
| `logging.level` | Log verbosity. Options: `debug`, `info`, `warn`, `error`. Default: `info`. |
| `logging.file` | Absolute path to the log file. |
