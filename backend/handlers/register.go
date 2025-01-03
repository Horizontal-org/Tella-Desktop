package handlers

import (
    "encoding/json"
    "net/http"
	
    "Tella-Desktop/backend/core/models"
    "Tella-Desktop/backend/core/ports"
)

type RegisterHandler struct {
    deviceService ports.DeviceService
}

func NewRegisterHandler(deviceService ports.DeviceService) *RegisterHandler {
    return &RegisterHandler{
        deviceService: deviceService,
    }
}

func (h *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var device models.Device
    if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.deviceService.RegisterDevice(&device); err != nil {
        http.Error(w, "Failed to register device", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    // Send empty response as per protocol
    json.NewEncoder(w).Encode(struct{}{})
}