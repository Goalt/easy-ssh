---
name: remote-ssh-executor
description: Connect to remote SSH servers (including over Cloudflare Tunnels) and execute shell commands safely. Trigger when connecting to remote hosts or running commands over SSH.
---

# Remote SSH Command Executor Skill

This skill provides standard procedures for connecting to remote SSH servers and executing commands cleanly, particularly using Cloudflare Tunnels (`cloudflared access tcp`) and configured server credentials.

## Server Credentials Reference
By default, check `.agents/AGENTS.md` for pre-configured server connection details.

Current Default Server:
- **Host**: `look-victory-listings-roland.trycloudflare.com`
- **User**: `osboxes`
- **Password**: `liverfag@P3`
- **Proxy Command**: `cloudflared access tcp --hostname %h`

---

## How to Execute Commands on Remote SSH Host

When executing commands on the remote SSH server, use standard SSH flags to avoid interactive prompts or hanging:

### 1. Execute a Single Non-Interactive Command
```bash
ssh -o "StrictHostKeyChecking=accept-new" -o "ProxyCommand=cloudflared access tcp --hostname %h" osboxes@<host> "<command>"
```

### 2. Passing Passwords Automatically (if sshpass is available or via batch script)
If `sshpass` is installed:
```bash
sshpass -p 'liverfag@P3' ssh -o "StrictHostKeyChecking=accept-new" -o "ProxyCommand=cloudflared access tcp --hostname %h" osboxes@<host> "<command>"
```

---

## Guidelines & Best Practices

1. **Avoid Stalling on Key Prompts**: Always pass `-o StrictHostKeyChecking=accept-new` (or `-o StrictHostKeyChecking=no`) for dynamic `trycloudflare.com` domain names.
2. **Non-Interactive Commands**: Wrap remote commands in quotes (e.g. `ssh ... "ls -la /var/log"`).
3. **Capture Output**: Ensure standard output and error output are captured and presented clearly to the user.
