package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPascalCase(t *testing.T) {
	assert.Equal(t, "AuthService", ToPascalCase("auth-service"))
	assert.Equal(t, "BillingService", ToPascalCase("billing_service"))
	assert.Equal(t, "MyProject", ToPascalCase("myProject"))
	assert.Equal(t, "Auth", ToPascalCase("auth"))
	assert.Equal(t, "UsEast1", ToPascalCase("us-east-1"))
	assert.Equal(t, "EuWest1", ToPascalCase("eu-west-1"))
	assert.Equal(t, "ApNortheast2", ToPascalCase("ap-northeast-2"))
	assert.Equal(t, "", ToPascalCase(""))
}

func TestNameFormater(t *testing.T) {
	nf := DefaultNameFormater("hot")
	assert.Equal(t, "hot-api", nf.Format("api"))

	list := nf.FormatList([]string{"a", "b"})
	assert.NotNil(t, list)
	assert.Len(t, *list, 2)
	assert.Equal(t, "a-hot", *(*list)[0])
	assert.Equal(t, "b-hot", *(*list)[1])
}
