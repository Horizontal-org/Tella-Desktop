package services

import (
    "context"
    "Tella-Desktop/backend/core/models"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type DeviceService struct {
    ctx     context.Context
    devices map[string]*models.Device
}

func NewDeviceService(ctx context.Context) *DeviceService {
    return &DeviceService{
        ctx:     ctx,
        devices: make(map[string]*models.Device),
    }
}

func (s *DeviceService) RegisterDevice(device *models.Device) error {
    s.devices[device.Fingerprint] = device
    
    // Notify UI about new device registration
    deviceInfo := device.Alias + " (" + device.DeviceModel + ")"
    runtime.EventsEmit(s.ctx, "device-registered", deviceInfo)
    
    return nil
}

func (s *DeviceService) GetDevice(fingerprint string) (*models.Device, error) {
    device, exists := s.devices[fingerprint]
    if !exists {
        return nil, nil
    }
    return device, nil
}