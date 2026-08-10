package tools

import "fmt"

type NameFormater struct {
	Prefix    string
	Delimiter string
}

func (nf NameFormater) Format(name string) string {
	return fmt.Sprintf("%s%s%s", nf.Prefix, nf.Delimiter, name)
}

func DefaultNameFormater(name string) NameFormater {
	return NameFormater{Prefix: name, Delimiter: "-"}
}
