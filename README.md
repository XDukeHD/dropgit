# DropGit

DropGit is a production-ready backup daemon written in Go. It monitors a configurable directory (like `~/Projects`), backs up projects into compressed `tar.gz` archives, maintains a changelog using SHA256 hashing, and stores metadata in a local SQLite database. 

## Features

- **Automated Backups:** Run backups periodically (`daily`, `weekly`, `monthly`, custom cron).
- **Compression:** High-performance concurrent `tar.gz` compression.
- **Change Tracking:** Keeps a SQLite database mapping file changes and SHA256 hashes.
- **Ignore Patterns:** Supports default ignores (`node_modules`, `.git`, etc.) and project-specific `.dropgitignore` files.
- **Daemon Mode:** Runs smoothly as a user-level `systemd` service.
- **Validation:** Verifies archive integrity after generation using SHA256.
- **Graceful Lifecycle:** Handles SIGTERM/SIGINT for clean shutdown and SIGHUP for hot config reloading.

## Prerequisites

- Go 1.22 or higher
- C compiler (for `mattn/go-sqlite3` CGO support)
- Linux (Ubuntu/Fedora) with `systemd`

## Installation

1. Clone or download the repository.
2. Run the make install command:
   ```bash
   make install
   ```
3. Install and enable the `systemd` service:
   ```bash
   make service-install
   ```

The daemon will now run in the background.

## Configuration

The default configuration file is created automatically at `~/.config/dropgit/config.yml`.

To modify the schedule, backup paths, or other options, edit this file and reload the daemon:
```bash
systemctl --user reload dropgit.service
```

## Management

- Check status: `systemctl --user status dropgit.service`
- View logs: `journalctl --user -u dropgit -f`
- Stop the service: `systemctl --user stop dropgit.service`
- Run a manual backup right now: `dropgit -once`

## Uninstallation

To remove the service and binaries completely:
```bash
make service-remove
make uninstall
```


## ToDo:
- [ ] Add automatic upload to cloud storage (Google Drive, Dropbox, etc.)