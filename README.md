# Tella Desktop

A desktop version of the Tella app made to share files offline via p2p. 

## Prerequisites

- Go 1.21 or later
- Node.js 20.11.1 or later
- Wails CLI (v2.8.0+)

To install Wails:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Development

1. Clone the repository
2. Install frontend dependencies:
```bash
cd frontend
npm install
```

3. Run in development mode:
```bash
wails dev
```

This will start both the backend server and frontend development server with hot reload.

## Building

To build a production version:

```bash
wails build
```

The built application will be in the `build/bin` directory.


## Protocol Support

The application implements the LocalSend v2 protocol with the following endpoints:

- `/api/localsend/v2/register` - Device registration
- `/api/localsend/v2/prepare-upload` - Prepare file transfer
- `/api/localsend/v2/upload` - File upload