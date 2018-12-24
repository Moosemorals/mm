package eveapi

import (
	"encoding/json"
	"testing"
)

func TestGetTypes(t *testing.T) {

	types, err := getTypes()

	if err != nil {
		t.Fatal(err)
	}

	if len(types) == 0 {
		t.Fatal("Types should have contents")
	}

	t.Logf("Got %d type(s)", len(types))
	t.Logf("Type %d is %+v", 15596, types[15596])
}

func TestGetAttributes(t *testing.T) {
	attr, err := getAttributes()

	if err != nil {
		t.Fatal(err)
	}

	if len(attr) == 0 {
		t.Fatal("Attributes should have content")
	}

	t.Logf("Got %d attribute(s)", len(attr))
	t.Logf("Attr %d is %+v", 2775, attr[2775])
}

func TestSetMaterials(t *testing.T) {
	types, err := getTypes()
	if err != nil {
		t.Fatal(err)
	}

	err = setTypeMaterials(types)

	if err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(types[195])
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Type %d is %s", 195, b)
}
