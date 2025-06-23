package model

// Medication struct.
// TODO: typically there are two layers of model objects: storage and business.
// i.e. storage services return storage objects ->
// business layer reads storage objects, perform logic, create business objects ->
// API layer converts business objects into API.
// On practice having the same objects for both storage and business is enough and removes a bunch of boiler code.
//
// More information on fields: internal/medication/service.go
type Medication struct {
	Id     string
	Name   string
	Dosage string
	Form   string
}
