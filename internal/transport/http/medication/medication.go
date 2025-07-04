package medication

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/chestnut42/test-medication/internal/medication"
	"github.com/chestnut42/test-medication/internal/model"
	"github.com/chestnut42/test-medication/internal/utils/logx"
)

type createMedicationService interface {
	CreateMedication(ctx context.Context, identity model.Identity, data model.MedicationData) (model.Medication, error)
}

type createMedicationOutput struct {
	Id      string `json:"id"`
	Version string `json:"version"`
	Name    string `json:"name"`
	Dosage  string `json:"dosage"`
	Form    string `json:"form"`
}

func CreateMedication(svc createMedicationService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logx.Logger(r.Context())

		id := r.PathValue("id")
		if err := validateId(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		logger = logger.With(slog.String("id", id))

		var req medicationDataInput
		if err := readJson(r, &req); err != nil {
			// Likely JSON format errors - quite safe to send it back to the client
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mData, err := req.toMedicationData()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		owner := getOwner(r)
		logger = logger.With(slog.String("owner", owner))

		respObject, err := svc.CreateMedication(r.Context(), model.Identity{
			Id:    id,
			Owner: owner,
		}, mData)
		if err != nil {
			logger.Error("svc.CreateMedication",
				slog.Any("error", err))
			if errors.Is(err, medication.ErrAlreadyExists) {
				// TODO: ideally we should read existing object and return 200 if it's equal (+ even 201 for the first creation, but highly debatable)
				// If it's an API for a partner, usually we can allow weaker RESTful contracts.
				// Of course it's up to discussion with partners.
				http.Error(w, "already exists", http.StatusConflict)
				return
			}
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(createMedicationOutput{
			Id:      id,
			Version: respObject.Version,
			Name:    respObject.Name,
			Dosage:  respObject.Dosage,
			Form:    string(respObject.Form),
		}); err != nil {
			logger.Error("svc.CreateMedication")
			return
		}

		// OK
	})
}
