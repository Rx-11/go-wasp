package registry

type Registry interface {
	SaveFunction(name string, data []byte) error
	GetFunction(name string) ([]byte, error)
	ListFunctions() ([]string, error)
}
