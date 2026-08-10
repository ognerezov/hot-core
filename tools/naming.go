package tools

import "fmt"

type NameFormater struct {
	prefix    string
	delimiter string
}

func (nf NameFormater) Format(name string) string {
	return fmt.Sprintf("%s%s%s", nf.prefix, nf.delimiter, name)
}
