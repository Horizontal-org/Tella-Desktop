# Tella Desktop

A desktop version of the Tella app made to share files offline via p2p. This application enables secure, encrypted file transfers between devices without relying on external servers, prioritizing privacy and security for sensitive data exchange.

## Prerequisites

- Go 1.24 or later
- Node.js 20.11.1 or later
- Wails CLI (v2.10.1+)

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

For Linux systems, you may need to use:

```bash
wails dev -tags webkit2_41
```

This will start both the backend server and frontend development server with hot reload.

## Building

To build a production version:

```bash
wails build
```

The built application will be in the `build/bin` directory.


## Protocol Support

The application implements the Tella P2P protocol with the following endpoints:

- Default Port: 53317 (user configurable if unavailable)
- `POST /api/v1/ping` - Initial handshake for manual connections
- `POST /api/v1/register` - Device registration with PIN authentication
- `POST /api/v1/prepare-upload` - Prepare file transfer session
- `PUT /api/v1/upload` - File upload with binary data

## Platform-Specific Notes

### macOS Code Signing

The application is configured for code signing on macOS for distribution outside the App Store:

- Uses Developer ID Application certificate for notarization
- Includes hardened runtime options for security
- Requires valid Apple Developer account for signing

To build a signed version for macOS:

- Update the identity in wails.json with your Developer ID
- Ensure you have a valid Developer ID Application certificate
- Run wails build - the app will be automatically signed during build

### Compatibility

- Mobile: Compatible with Tella iOS and Android apps using the same P2P protocol
- Network: Requires devices to be on the same local network (Wi-Fi)