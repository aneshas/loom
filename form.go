package loom

type ViewModel struct {
	Values map[string]string // map[string]any ?
	Errors map[string]string
}

func (vm *ViewModel) HasErrors() bool {
	return len(vm.Errors) > 0
}
