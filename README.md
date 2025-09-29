# Tella Desktop

A desktop version of the Tella app made to share files offline via Nearby Sharing. This application enables secure, encrypted file transfers between devices without relying on external servers, prioritizing privacy and security for sensitive data exchange. 

## Platform and availability
Nearby Sharing will be available for Tella Android, Tella iOS and Tella Desktop, but it's still under development.

The feature is still in alpha, and it's currently being audited by an independent security firm. It will be launched to production only after the priority security fixes are implemented.

User facing documentation: 
- about the Nearby Sharing in general: https://beta.tella-app.org/nearby-sharing
- about Tella Desktop: https://beta.tella-app.org/get-started-desktop/



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

### Building for Windows

We can leverage the Zig toolchain to easily cross-compile golang projects with CGO
dependencies, as demonstrated in [article 1](https://infinitedigits.co/tinker/go-and-zig/),
[article 2](https://archive.is/6zlX8) and combine that with Wails [manual
builds](https://wails.io/docs/guides/manual-builds/#manual-steps-3).

* Install the [Zig compiler](https://wiki.archlinux.org/title/Zig)
* Run the script `build-for-windows.sh`

The Windows executable will be saved as `tella.exe`.

`build-for-windows.sh`
```sh
#!/bin/bash
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="zig cc -target x86_64-windows" CXX="zig cc -target x86_64-windows" go build -tags desktop,production -ldflags "-w -s -H windowsgui" -o tella.exe
```

## Protocol Support

The application implements the [Tella Nearby Sharing protocol](https://github.com/Horizontal-org/Tella-P2P-Protocol) with the following endpoints:

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
