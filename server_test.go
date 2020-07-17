package easyHttpCrud

import (
	"testing"
)

func TestCreateRouter(t *testing.T) {
	type Model struct {
		Id   int
		Name string
	}
	Serve(&Model{}, map[string]interface{}{})
}
