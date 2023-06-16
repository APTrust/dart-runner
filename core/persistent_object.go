package core

type PersistentObject interface {
	Delete() error
	GetErrors() map[string]string
	ObjID() string
	ObjName() string
	ObjType() string
	Save() error
	String() string
	ToForm() *Form
	Validate() bool
}
