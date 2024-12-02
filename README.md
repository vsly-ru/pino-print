# pino-print

A lightweight CLI tool that pretty prints [Pino](https://getpino.io/) JSON logs with colors and formatting.

## Features

- Reads Pino JSON logs from stdin
- Color-coded log levels
- Highlighted fields of additional data
- Zero dependencies (doesn't even need nodejs or npm)


## Installation

### Prerequisites

- Go 1.23 or higher

### Install from source

Clone the repository and run the build script:
```bash
git clone https://github.com/vsly-ru/pino-print.git
cd pino-print
chmod +x scripts/install.sh
./scripts/install.sh
```

This will compile and install the binary to `/usr/local/bin`.

## Usage

Pipe your Pino logs to `pino-print`:
```bash
bash
node app.js | pino-print
```
Or with an existing log file:
```bash
cat logs.json | pino-print
```


## Color Scheme

- FATAL/ERROR: Red
- WARN: Yellow
- INFO: Green
- DEBUG: Blue
- TRACE: Grey

## License

MIT

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.