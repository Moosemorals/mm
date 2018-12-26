package eveapi

import (
	"encoding/json"
	"testing"
)

func TestGetTypes(t *testing.T) {

	eve := &Eve{}
	err := eve.loadTypes()

	if err != nil {
		t.Fatal(err)
	}

	if len(eve.types) == 0 {
		t.Fatal("Types should have contents")
	}

	t.Logf("Got %d type(s)", len(eve.types))
	t.Logf("Type %d is %+v", 15596, eve.types[15596])
}

func TestGetAttributes(t *testing.T) {
	eve := &Eve{}
	err := eve.loadAttributes()

	if err != nil {
		t.Fatal(err)
	}

	if len(eve.attributes) == 0 {
		t.Fatal("Attributes should have content")
	}

	t.Logf("Got %d attribute(s)", len(eve.attributes))
	t.Logf("Attr %d is %+v", 2775, eve.attributes[2775])
}

func TestSetMaterials(t *testing.T) {
	eve := &Eve{}

	err := eve.loadTypes()
	if err != nil {
		t.Fatal(err)
	}

	err = eve.loadTypeMaterials()

	if err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(eve.types[195])
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Type %d is %s", 195, b)
}

func TestReadBlueprints(t *testing.T) {
	eve := &Eve{}

	if err := eve.loadTypes(); err != nil {
		t.Fatal(err)
	}

	err := eve.loadBlueprints()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Blueprint %d is %+v", 1002, eve.blueprints[1002])
	b, err := json.Marshal(eve.blueprints[1002])
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Blueprint %d is %s", 1002, b)

	b2, err := json.Marshal(eve.types[195])
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("type %d is %s", 195, b2)
}
