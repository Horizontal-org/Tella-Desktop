package registration

import (
    "encoding/json"
    "net/http"
)

type Handler struct {
    service Service
}


func NewHandler(service Service) *Handler {
    return &Handler{
        service: service,
    }
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var device Device
    if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.service.Register(&device); err != nil {
        http.Error(w, "Failed to register device", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(struct{}{})
}