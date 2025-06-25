package medication

import (
	"net/http"
)

type getMedicationService interface {
}

func GetMedication(svc getMedicationService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	})
}
