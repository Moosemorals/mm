package eveapi

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

func getTypes() (eveTypes, error) {
	result := make(map[int32]*EveType)

	f, err := os.Open("../wwwroot/eve/data/fsd/typeIDs.json")
	if err != nil {
		return result, err
	}

	dec := json.NewDecoder(f)

	// Read start of hash
	t, err := dec.Token()
	if err != nil {
		return result, err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return result, errors.New("File should start with an {")
	}

	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return result, err
		}
		itemID, err := strconv.Atoi(t.(string))
		if err != nil {
			return result, err
		}

		var e EveType
		if err := dec.Decode(&e); err != nil {
			return result, err
		}
		if e.Published {
			result[int32(itemID)] = &e
		}

	}
	return result, nil
}

func getAttributes() (eveAttributes, error) {
	result := make(map[int32]*EveAttribute)

	f, err := os.Open("../wwwroot/eve/data/bsd/dgmAttributeTypes.json")
	if err != nil {
		return result, err
	}

	dec := json.NewDecoder(f)

	// Read start of array
	t, err := dec.Token()
	if err != nil {
		return result, err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return result, errors.New("File should start with an [")
	}

	for dec.More() {
		var a EveAttribute
		if err := dec.Decode(&a); err != nil {
			return result, err
		}

		if a.Published {
			result[a.ID] = &a
		}
	}

	return result, nil
}

func setTypeAttributes(types eveTypes, attr eveAttributes) error {

	type Tuple struct {
		AttrID     int32    `json:"attributeID"`
		TypeID     int32    `json:"typeID"`
		ValueInt   *int64   `json:"valueInt"`
		ValueFloat *float64 `json:"valueFloat"`
	}

	f, err := os.Open("../wwwroot/eve/data/bsd/dgmTypeAttributes.json")
	if err != nil {
		return err
	}

	dec := json.NewDecoder(f)

	// Read start of array
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return errors.New("File should start with an [")
	}

	for dec.More() {
		var p Tuple
		if err := dec.Decode(&p); err != nil {
			return err
		}

		t, ok := types[p.TypeID]
		if !ok {
			continue
		}
		a, ok := attr[p.AttrID]
		if !ok {
			continue
		}

		if t.Attributes == nil {
			t.Attributes = []TypeAttr{}
		}

		var v float64
		if p.ValueFloat != nil {
			v = *p.ValueFloat
		} else {
			v = float64(*p.ValueInt)
		}

		t.Attributes = append(t.Attributes, TypeAttr{
			Attribute: a,
			Value:     v,
		})

	}
	return nil
}

func setTypeMaterials(types eveTypes) error {

	type Tuple struct {
		MaterialTypeID int32 `json:"materialTypeId"`
		TypeID         int32 `json:"typeID"`
		Quantity       int64 `json:"quantity"`
	}

	f, err := os.Open("../wwwroot/eve/data/bsd/invTypeMaterials.json")
	if err != nil {
		return err
	}

	dec := json.NewDecoder(f)

	// Read start of array
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return errors.New("File should start with an [")
	}

	for dec.More() {
		var p Tuple
		if err := dec.Decode(&p); err != nil {
			return err
		}

		t, ok := types[p.TypeID]
		if !ok {
			continue
		}

		if t.Materials == nil {
			t.Materials = []TypeValue{}
		}
		t.Materials = append(t.Materials, TypeValue{
			ID:    p.MaterialTypeID,
			Value: p.Quantity,
		})

	}
	return nil
}
