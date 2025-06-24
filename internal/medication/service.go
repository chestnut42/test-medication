package medication

import (
	"context"
	"errors"
	"fmt"

	"github.com/chestnut42/test-medication/internal/model"
	"github.com/chestnut42/test-medication/internal/storage"
)

// This is a package for business logic. Given the simplicity of the task it could have been omitted.
// Why it's here:
// 1. Request validation is to be done by API layer. However, that only should be simple validation, like the length
//    of the string/array, possible values if that's enum or some range for the number. More complex validation can
//    require additional calls to dependencies or DB.
// 2. Name of the drug: very likely it must be validated so that we don't have multiple paracetamol versions in DB
//    (like "Paracetamol" and "paracetamol", and bet there's going to be "Para Cetamol" on top). Since there are many
//    drugs out there, complete medications list is going to be big. It will be a whole separate service for that or
//    at least a separate call to DB. + some case/whitespace insensitive logic. This kind of validation should be
//    a part of business layer. Hence TODO: validate drug name
// 3. Dosage. Perhaps our application should not be aware of what's written inside Dosage field. For example,
//    we just take what the doctor has written and repeat it back. That can be the case if all doctors are
//    really write what they like and there's little standardisation in that regard. However, very likely it's not
//    the case. We should have at least {number: int, unit: string}. So that 500mg is {number: 500, unit: "mg"}.
//    On top of it - units should be validated against form and the drug, e.g. given it's Paracetamol in Tablet
//    the unit can't be "ml". TODO: validate and structure the dosage
// 4. Very likely, form is a short enum. It should be confirmed, of course. But it's very likely enum, thus I'm
//    implementing an enum here.
// 5. We have Version<>Conflict logic for updates. I'd rather have it here instead of adding complexity to storage layer.

type Storage interface {
	CreateMedication(ctx context.Context, medication model.Medication) error
}

type NewVersionFunc func() string

type Service struct {
	store Storage

	newVersion NewVersionFunc
}

func NewService(store Storage) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) CreateMedication(ctx context.Context, identity model.Identity, data model.MedicationData) (model.Medication, error) {
	if identity.Owner == "" {
		// It's not a BadInput. Caller of this function must ensure the owner is determined and is not empty.
		return model.Medication{}, errors.New("owner is required")
	}

	// TODO: Validate name and dosage taking form into account

	storedMedication := model.Medication{
		Identity:       identity,
		Version:        s.newVersion(),
		MedicationData: data,
	}
	if err := s.store.CreateMedication(ctx, storedMedication); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return model.Medication{}, fmt.Errorf("medication %v already exists: %w", identity, ErrAlreadyExists)
		}
		return model.Medication{}, fmt.Errorf("creating medication: %w", err)
	}
	return storedMedication, nil
}
