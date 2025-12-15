### build

```bash
go build ./cmd/go-scanner
```

### usage (TCP connect scan)

```bash
go-scanner.exe tcp connect -p 1-500 google.com
```

### Flags

#### `-p`

Ports to scan.

Supported formats:

- Single port: `80`
- Range: `1-1024`
- List: `22,80,443`

#### `--profile`

Scan profile preset (default: `default`)

Available profiles:

- `passive`: Passive scan, no active probing (timeout: 2s, concurrency: 50)
- `default`: Balanced scan, service detection only (timeout: 1s, concurrency: 100)
- `aggressive`: Fast scan with active HTTP/HTTPS probing (timeout: 500ms, concurrency: 200)

```bash
go-scanner.exe tcp connect --profile passive -p 22,80,443 scanme.nmap.org
```

#### `--banner`

Enable passive banner grabbing on supported ports (FTP, SSH, SMTP, POP3, IMAP).

#### `--probe`

Enable active probing on detected services. Can override profile settings.

```bash
# Use passive profile but enable active probing
go-scanner.exe tcp connect --profile passive --probe -p 80,443 example.com
```

#### `--probe-types`

Comma-separated list of probe types to run (default: http,https)

#### `--timeout`

Timeout per connection in milliseconds. Overrides profile default.

```bash
go-scanner.exe tcp connect --profile aggressive --timeout 2000 -p 1-1000 target.com
```

#### `--threads`

Maximum number of concurrent connections. Overrides profile default.

### Examples

```bash
# Default profile (implicit)
go-scanner.exe tcp connect -p 80,443 google.com

# Passive reconnaissance
go-scanner.exe tcp connect --profile passive -p 1-1000 scanme.nmap.org

# Aggressive scan with active probing
go-scanner.exe tcp connect --profile aggressive -p 80,443,8080 target.com

# Passive profile with active probing override
go-scanner.exe tcp connect --profile passive --probe --banner -p 22,80,443 target.com
```
