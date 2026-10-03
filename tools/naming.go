package tools

import (
	"fmt"
	"strings"
	"unicode"

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

// ToPascalCase converts a hyphen, underscore, or space separated string into PascalCase (UpperCamelCase).
func ToPascalCase(s string) string {
	var parts []string
	for _, p := range strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	}) {
		if len(p) > 0 {
			r := []rune(p)
			r[0] = unicode.ToUpper(r[0])
			parts = append(parts, string(r))
		}
	}
	return strings.Join(parts, "")
}
