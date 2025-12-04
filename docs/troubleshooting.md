# Troubleshooting

## Common Issues

### Agent won't start
If the agent fails to start, first run the validation command to identify potential issues:
```bash
fixpanic agent validate
```
Then, check the logs for specific error messages:
```bash
fixpanic agent logs
```

### Connectivity Problems
The agent requires a persistent connection to the FixPanic platform.
- **Check Network**: Ensure your server can reach `socket.fixpanic.com` on port `9000`.
  ```bash
  curl -I socket.fixpanic.com:9000
  ```
- **Firewalls/Proxies**: If you are behind a strict firewall or corporate proxy, ensure the traffic is allowed.

### Permission Errors
- **System Install**: If you installed as root (using `sudo`), you must run management commands (start, stop, uninstall) with `sudo`.
- **User Install**: If you installed as a regular user, do **not** use `sudo` for agent commands, as this may cause permission conflicts with config files in your home directory.

## Support Channels

If you cannot resolve the issue, please contact us:
- 📧 **Email**: [support@fixpanic.com](mailto:support@fixpanic.com)
- 📖 **Documentation**: [docs.fixpanic.com](https://docs.fixpanic.com)
- 🐛 **GitHub Issues**: [fixpanic-cli-tool/issues](https://github.com/fixpanic/fixpanic-cli-tool/issues)
- 💬 **Discord Community**: [Join Discord](https://discord.gg/fixpanic)
