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

#### `--banner`

Enable passive banner grabbing on supported ports.

#### `--probe`

Enable active probing on detected services.

```bash
go-scanner.exe tcp connect --banner --probe -p 80 example.com
```

#### `--probe-types`

Comma-separated list of probe types to run (default: http,https)
