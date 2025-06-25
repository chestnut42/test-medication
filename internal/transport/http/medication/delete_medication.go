package medication

import (
	"net/http"
)

type deleteMedicationService interface {
}

func DeleteMedication(svc deleteMedicationService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	})
}
