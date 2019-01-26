package validator

type validatable interface {
	Validate() (bool, []error)
}

type Validator struct {
	errs []error
}

func NewValidator() *Validator {
	return &Validator{
		errs: []error{},
	}
}

func (v *Validator) AddError(err ...error) {
	v.errs = append(v.errs, err...)
}

func (v *Validator) Validate(obj validatable) {
	if ok, err := obj.Validate(); !ok {
		v.AddError(err...)
	}
}

func (v *Validator) Valid() (bool, []error) {
	if len(v.errs) > 0 {
		return false, v.errs
	}
	return true, nil
}
