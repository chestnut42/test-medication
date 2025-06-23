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
	Dosage string // It shall likely be a struct. See internal/medication/service.go for details
	Form   string // It can be enum, but I'd prefer the business logic layer to validate it. Deep down the line enum shoud be a string in DB (not number)
}
