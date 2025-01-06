package registration

import (
    "context"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type service struct {
    ctx     context.Context
    peers   map[string]*Device
}

func NewService(ctx context.Context) Service {
    return &service{
        ctx:     ctx,
        peers:   make(map[string]*Device),
    }
}

func (s *service) Register(device *Device) error {
    s.peers[device.Fingerprint] = device
    
    // Notify UI about new device registration
    deviceInfo := device.Alias + " (" + device.DeviceModel + ")"
    runtime.EventsEmit(s.ctx, "device-registered", deviceInfo)
    
    return nil
}
