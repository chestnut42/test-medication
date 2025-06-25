package medication

import (
	"net/http"
)

type updateMedicationService interface {
}

func UpdateMedication(svc updateMedicationService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	})
}
