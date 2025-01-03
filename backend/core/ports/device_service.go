package ports

import (
    "Tella-Desktop/backend/core/models"
)

type DeviceService interface {
    RegisterDevice(device *models.Device) error
    GetDevice(fingerprint string) (*models.Device, error)
}
