package constants

import "errors"

var ErrObjecValidation = errors.New("object contains validation errors")

var ErrNotDeletable = errors.New("object cannot be deleted because it is required by dart")

var ErrInvalidOperation = errors.New("this operation is invalid or forbidden")

var ErrUnknownType = errors.New("unknown type")

var ErrWrongTypeForForm = errors.New("cannot create form from result: query result is either a list or empty")

var ErrUniqueConstraint = errors.New("An object of this type already exists with the same name. Choose a new name.")
