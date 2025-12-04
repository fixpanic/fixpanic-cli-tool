# Commands Reference

## `fixpanic agent install`
Installs and configures a new agent.

**Usage:**
```bash
fixpanic agent install --agent-id=<id> --api-key=<key>
```

**Flags:**
- `--agent-id`: (Required) The unique ID for the agent.
- `--api-key`: (Required) The authentication key.

---

## `fixpanic agent status`
Checks the current status of the agent service.

**Usage:**
```bash
fixpanic agent status
```

**Output:**
- Running state (Active/Stopped)
- Connection status (Connected/Disconnected)
- Version information

---

## `fixpanic agent start`
Manually starts the agent process.

**Usage:**
```bash
fixpanic agent start
```

---

## `fixpanic agent stop`
Manually stops the agent process.

**Usage:**
```bash
fixpanic agent stop
```

---

## `fixpanic agent logs`
Displays the agent's log output.

**Usage:**
```bash
fixpanic agent logs [--follow] [--lines=100]
```

**Flags:**
- `--follow` (`-f`): Stream logs in real-time.
- `--lines` (`-n`): Number of lines to show (default: 100).

---

## `fixpanic agent validate`
Validates the installation and configuration. Checks for permission issues, missing files, or invalid config.

**Usage:**
```bash
fixpanic agent validate
```

---

## `fixpanic agent uninstall`
Removes the agent, configuration, and logs.

**Usage:**
```bash
fixpanic agent uninstall [--force]
```

**Flags:**
- `--force`: Remove without confirmation prompt.

---

## `fixpanic upgrade`
Updates the CLI tool to the latest available version.

**Usage:**
```bash
fixpanic upgrade
```
