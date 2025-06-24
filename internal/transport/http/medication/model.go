package medication

import (
	"errors"
	"fmt"

	"github.com/chestnut42/test-medication/internal/model"
)

type medicationDataInput struct {
	Name   string `json:"name"`
	Dosage string `json:"dosage"`
	Form   string `json:"form"`
}

func (cmi medicationDataInput) toMedicationData() (model.MedicationData, error) {
	if cmi.Name == "" {
		return model.MedicationData{}, errors.New("name must not be empty")
	}
	if len(cmi.Name) >= 1024 {
		return model.MedicationData{}, errors.New("name must be less than 1024 characters")
	}

	if cmi.Dosage == "" {
		return model.MedicationData{}, errors.New("dosage must not be empty")
	}
	if len(cmi.Dosage) >= 1024 {
		return model.MedicationData{}, errors.New("dosage must be less than 1024 characters")
	}

	parsedForm, ok := model.ParseForm(cmi.Form)
	if !ok {
		return model.MedicationData{}, fmt.Errorf("<%s> is not a valid form", cmi.Form)
	}

	return model.MedicationData{
		Name:   cmi.Name,
		Dosage: cmi.Dosage,
		Form:   parsedForm,
	}, nil
}

func validateId(id string) error {
	if id == "" {
		return errors.New("id must not be empty")
	}
	if len(id) >= 64 {
		return errors.New("id must be less than 64 characters")
	}
	return nil
}
