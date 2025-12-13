### build

```bash
go build ./cmd/go-scanner
```

### usage (TCP connect scan)

```bash
go-scanner.exe tcp connect -p 1-500 google.com
```

#### banner grabbing

with the flag "--banner", the scanner will try to grab the banner of the service running on the port :3
