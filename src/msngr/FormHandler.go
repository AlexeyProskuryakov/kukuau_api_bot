package msngr

//experiments for refactoring...

type FormHandler struct {
	form OutForm
}

func (fh FormHandler) validate(form InForm) (s string, e error) {
	return
}

func (fh FormHandler) getValue(key string) (val *interface{}) {
	return
}
