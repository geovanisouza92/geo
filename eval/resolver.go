package eval

import (
	"io/ioutil"
)

type ModuleResolver interface {
	Resolve(string) string
}

type FileResolver struct {
}

func (fr *FileResolver) Resolve(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		// TODO: fmt.Print(err.Error())
		return ""
	}
	return string(b)
}
