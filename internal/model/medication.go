package model

import (
	"strings"
)

// Medication struct.
// TODO: typically there are two layers of model objects: storage and business.
// i.e. storage services return storage objects ->
// business layer reads storage objects, perform logic, create business objects ->
// API layer converts business objects into API.
// On practice having the same objects for both storage and business is enough and removes a bunch of boiler code.
//
// More information on fields: internal/medication/service.go
type Medication struct {
	Identity
	MedicationData
	Version string
}

type Identity struct {
	Id    string
	Owner string
}

type MedicationData struct {
	Name   string
	Dosage string // It shall likely be a struct. See internal/medication/service.go for details
	Form   Form   // It's important to save the string to DB. Validation happens on API/Business layer
}

type Form string

const (
	FormTablet  Form = "tablet"
	FormCapsule Form = "capsule"
	FormLiquid  Form = "liquid"
)

func ParseForm(form string) (Form, bool) {
	for _, f := range []Form{FormTablet, FormCapsule, FormLiquid} {
		if string(f) == form {
			return f, true
		}
		if strings.ToLower(strings.TrimSpace(form)) == strings.ToLower(string(f)) {
			return f, true
		}
	}
	return "", false
}
