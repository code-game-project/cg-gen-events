package lang

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-game-project/cg-gen-events/cge"
)

type jsonObject struct {
	GameName   string      `json:"game_name"`
	CGEVersion string      `json:"cge_version"`
	Comments   []string    `json:"comments,omitempty"`
	Events     []jsonEvent `json:"events"`
	Types      []jsonType  `json:"types"`
	Enums      []jsonEnum  `json:"enums"`
}

type jsonEvent struct {
	Name       string         `json:"name"`
	Comments   []string       `json:"comments,omitempty"`
	Properties []jsonProperty `json:"properties"`
}

type jsonType struct {
	Name       string         `json:"name"`
	Comments   []string       `json:"comments,omitempty"`
	Properties []jsonProperty `json:"properties"`
}

type jsonProperty struct {
	Name     string           `json:"name"`
	Comments []string         `json:"comments,omitempty"`
	Type     jsonPropertyType `json:"type"`
}

type jsonPropertyType struct {
	Name    string            `json:"name"`
	Generic *jsonPropertyType `json:"generic,omitempty"`
}

type jsonEnum struct {
	Name     string          `json:"name"`
	Comments []string        `json:"comments,omitempty"`
	Values   []jsonEnumValue `json:"values"`
}

type jsonEnumValue struct {
	Name     string   `json:"name"`
	Comments []string `json:"comments,omitempty"`
}

type JSON struct {
	builder strings.Builder
	json    jsonObject
}

func (j *JSON) Generate(metadata cge.Metadata, objects []cge.Object, dir string) error {
	file, err := os.Create(filepath.Join(dir, "events.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	j.builder = strings.Builder{}

	j.json = jsonObject{
		GameName:   metadata.Name,
		CGEVersion: metadata.CGEVersion,
		Comments:   metadata.Comments,
		Events:     make([]jsonEvent, 0, len(objects)/3),
		Types:      make([]jsonType, 0, len(objects)/3),
		Enums:      make([]jsonEnum, 0, len(objects)/3),
	}

	for _, object := range objects {
		if object.Type == cge.EVENT {
			j.generateEvent(object)
		} else if object.Type == cge.TYPE {
			j.generateType(object)
		} else {
			j.generateEnum(object)
		}
	}

	return json.NewEncoder(file).Encode(j.json)
}

func (j *JSON) generateEvent(object cge.Object) {
	j.json.Events = append(j.json.Events, jsonEvent{
		Name:       object.Name,
		Comments:   object.Comments,
		Properties: j.generateProperties(object.Properties),
	})
}

func (j *JSON) generateType(object cge.Object) {
	j.json.Types = append(j.json.Types, jsonType{
		Name:       object.Name,
		Comments:   object.Comments,
		Properties: j.generateProperties(object.Properties),
	})
}

func (j *JSON) generateEnum(object cge.Object) {
	j.json.Enums = append(j.json.Enums, jsonEnum{
		Name:     object.Name,
		Comments: object.Comments,
		Values:   j.generateEnumValues(object.Properties),
	})
}

func (j *JSON) generateProperties(properties []cge.Property) []jsonProperty {
	props := make([]jsonProperty, len(properties))
	for i, p := range properties {
		props[i] = jsonProperty{
			Name:     p.Name,
			Comments: p.Comments,
			Type:     *j.generatePropertyType(p.Type),
		}
	}
	return props
}

func (j *JSON) generatePropertyType(propertyType *cge.PropertyType) *jsonPropertyType {
	t := &jsonPropertyType{
		Name: strings.ToLower(string(propertyType.Token.Type)),
	}

	if propertyType.Token.Type == cge.IDENTIFIER {
		t.Name = propertyType.Token.Lexeme
	}

	if propertyType.Generic != nil {
		t.Generic = j.generatePropertyType(propertyType.Generic)
	}

	return t
}

func (j *JSON) generateEnumValues(properties []cge.Property) []jsonEnumValue {
	values := make([]jsonEnumValue, len(properties))
	for i, p := range properties {
		values[i] = jsonEnumValue{
			Name:     p.Name,
			Comments: p.Comments,
		}
	}
	return values
}
