# Commands Reference

## `opssquad node install`
Installs and configures a new node.

**Usage:**
```bash
opssquad node install --node-id=<id> --token=<token>
```

**Flags:**
- `--node-id`: (Required) The unique ID for the node.
- `--token`: (Required) The authentication token.
- `--force`: Force reinstall even if node is already installed.

---

## `opssquad node status`
Checks the current status of the node service.

**Usage:**
```bash
opssquad node status
```

**Output:**
- Running state (Active/Stopped)
- Connection status (Connected/Disconnected)
- Version information

---

## `opssquad node start`
Manually starts the node process.

**Usage:**
```bash
opssquad node start
```

---

## `opssquad node stop`
Manually stops the node process.

**Usage:**
```bash
opssquad node stop
```

---

## `opssquad node logs`
Displays the node's log output.

**Usage:**
```bash
opssquad node logs [--follow] [--lines=100]
```

**Flags:**
- `--follow` (`-f`): Stream logs in real-time.
- `--lines` (`-n`): Number of lines to show (default: 100).

---

## `opssquad node validate`
Validates the installation and configuration. Checks for permission issues, missing files, or invalid config.

**Usage:**
```bash
opssquad node validate
```

---

## `opssquad node uninstall`
Removes the node, configuration, and logs.

**Usage:**
```bash
opssquad node uninstall [--force]
```

**Flags:**
- `--force`: Remove without confirmation prompt.

---

## `opssquad upgrade`
Updates the CLI tool to the latest available version.

**Usage:**
```bash
opssquad upgrade
```
