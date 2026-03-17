package model

// Vessel represents a vessel related to booking
type Vessel struct {
	Name            string `json:"name"`
	VesselIMONumber string `json:"vesselIMONumber,omitempty"`
}
