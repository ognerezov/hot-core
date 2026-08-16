package tools

import (
	"fmt"

	"github.com/aws/jsii-runtime-go"
)

type NameFormater struct {
	Prefix    string
	Delimiter string
}

func (nf NameFormater) Format(name string) string {
	return fmt.Sprintf("%s%s%s", nf.Prefix, nf.Delimiter, name)
}

func (nf NameFormater) FormatList(list []string) *[]*string {
	return FormatList(list, fmt.Sprintf("%s%s", nf.Delimiter, nf.Prefix))
}

func DefaultNameFormater(name string) NameFormater {
	return NameFormater{Prefix: name, Delimiter: "-"}
}

func FormatList(list []string, suffix string) *[]*string {

	res := Map(list, func(t string) *string {
		return jsii.String(t + suffix)
	})
	return &res
}
