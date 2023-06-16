package core

type PersistentObject interface {
	ObjID() string
	ObjName() string
	ObjType() string
	Save() error
	Delete() error
	ToForm() *Form
	Validate() bool
	String() string
}
