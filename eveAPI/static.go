package eveapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (e *Eve) loadTypes() error {
	e.types = make(map[int32]*EveType)

	f, err := os.Open("../wwwroot/eve/data/fsd/typeIDs.json")
	if err != nil {
		return err
	}

	dec := json.NewDecoder(f)

	// Read start of hash
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return errors.New("File should start with an {")
	}

	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return err
		}
		itemID, err := strconv.Atoi(t.(string))
		if err != nil {
			return err
		}

		var x EveType
		if err := dec.Decode(&x); err != nil {
			return err
		}
		if x.Published {
			e.types[int32(itemID)] = &x
		}

	}
	return nil
}

func (e *Eve) loadBlueprints() error {
	e.blueprints = eveBlueprints{}

	f, err := os.Open("../wwwroot/eve/data/fsd/blueprints.json")
	if err != nil {
		return err
	}

	if err := json.NewDecoder(f).Decode(&e.blueprints); err != nil {
		return err
	}

	for _, bp := range e.blueprints {
		m, ok := bp.Activities["manufacturing"]
		if !ok {
			continue
		}

		p := m.Products
		if len(p) == 0 {
			continue
		}

		for _, x := range p {
			t, ok := e.types[x.TypeID]
			if ok {
				t.Blueprints = append(t.Blueprints, bp)
			}
		}
	}

	return nil

}

func (e *Eve) loadAttributes() error {
	e.attributes = make(map[int32]*EveAttribute)

	f, err := os.Open("../wwwroot/eve/data/bsd/dgmAttributeTypes.json")
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
		var a EveAttribute
		if err := dec.Decode(&a); err != nil {
			return err
		}

		if a.Published {
			e.attributes[a.ID] = &a
		}
	}

	return nil
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

func (e *Eve) loadGroups() error {
	f, err := os.Open("../wwwroot/eve/data/fsd/groupIds.json")
	if err != nil {
		return err
	}

	if err := json.NewDecoder(f).Decode(&e.groups); err != nil {
		return err
	}

	return nil
}

func (e *Eve) loadTypeMaterials() error {

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

		t, ok := e.types[p.TypeID]
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

func (e *Eve) loadStatic() error {
	log.Println("Loading types")
	if err := e.loadTypes(); err != nil {
		return err
	}

	log.Println("Loading materials")
	if err := e.loadTypeMaterials(); err != nil {
		return err
	}

	log.Println("Loading blueprints")
	if err := e.loadBlueprints(); err != nil {
		return err
	}

	log.Println("Loading Groups")
	if err := e.loadGroups(); err != nil {
		return err
	}

	return nil
}

func simpleTypeFromType(t *EveType) simpleType {
	return simpleType{
		Name:        t.Name["en"],
		Description: t.Description["en"],
		PortionSize: t.PortionSize,
		Materials:   t.Materials,
		Blueprints:  t.Blueprints,
	}
}

func (e *Eve) getTypesByID(ids []int32) map[int32]simpleType {
	result := make(map[int32]simpleType)

	mats := make(map[int32]bool)

	for _, id := range ids {
		t, ok := e.types[id]
		if ok {
			result[id] = simpleTypeFromType(t)
			for _, m := range t.Materials {
				mats[m.ID] = true
			}
		}
	}

	for matID := range mats {
		_, ok := result[matID]
		if !ok {
			t, ok2 := e.types[matID]
			if ok2 {
				result[matID] = simpleTypeFromType(t)
			}
		}
	}
	return result
}

func (e *Eve) handleGetTypesByID(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		writeError(w, 405, "Method not allowed", nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeError(w, 400, "Bad form data", err)
		return
	}

	rawIDs := strings.Split(r.PostForm.Get("ids"), ",")
	if len(rawIDs) == 0 {
		writeError(w, 400, "Missing id list", nil)
		return
	}

	ids := []int32{}
	for _, r := range rawIDs {
		rawID, err := strconv.Atoi(r)
		if err != nil {
			writeError(w, 400, "Bad id", err)
			return
		}

		ids = append(ids, int32(rawID))
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)

	json.NewEncoder(w).Encode(e.getTypesByID(ids))
}

func (e *Eve) getBuildables(w http.ResponseWriter, r *http.Request) {
	ids := []int32{}

	for id, t := range e.types {
		g := e.groups[t.GroupID]
		if g.CategoryID == 7 {
			ids = append(ids, id)
		}
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)

	json.NewEncoder(w).Encode(e.getTypesByID(ids))

}

func (e *Eve) handleStatic(w http.ResponseWriter, r *http.Request) {

	if strings.HasPrefix(r.URL.Path, "/eveapi/static/typesById") {
		e.handleGetTypesByID(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/eveapi/static/buildables") {
		e.getBuildables(w, r)
	}
}
