package core

type PersistentObject interface {
	GetErrors() map[string]string
	IsDeletable() bool
	ObjID() string
	ObjName() string
	ObjType() string
	String() string
	ToForm() *Form
	Validate() bool
}
