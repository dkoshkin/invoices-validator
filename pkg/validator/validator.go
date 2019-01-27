package validator

type ValidationError struct {
	Actual         string
	Expected       string
	AdditionalInfo string
}

type validatable interface {
	Validate() (bool, []ValidationError)
}

type Validator struct {
	errs []ValidationError
}

func NewValidator() *Validator {
	return &Validator{
		errs: []ValidationError{},
	}
}

func (v *Validator) AddError(err ...ValidationError) {
	v.errs = append(v.errs, err...)
}

func (v *Validator) Validate(obj validatable) {
	if ok, err := obj.Validate(); !ok {
		v.AddError(err...)
	}
}

func (v *Validator) Valid() (bool, []ValidationError) {
	if len(v.errs) > 0 {
		return false, v.errs
	}
	return true, nil
}
