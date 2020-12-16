package model

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/jinzhu/copier"
	"github.com/mitchellh/go-wordwrap"
)

// Model represents a general table of the JW Library database and
// defines methods that are needed so entries are mergeable.
type Model interface {
	ID() int
	SetID(int)
	UniqueKey() string
	Equals(m2 Model) bool
	RelatedEntries(db *Database) Related
	PrettyPrint(db *Database) string
	tableName() string
	idName() string
	scanRow(row *sql.Rows) (Model, error)
}

// Related combines entries that are related to a given model
type Related struct {
	BlockRange          []*BlockRange       `json:"blockRange"`
	Bookmark            *Bookmark           `json:"bookmark"`
	Location            *Location           `json:"location"`
	PublicationLocation *Location           `json:"publicationLocation"`
	Note                *Note               `json:"note"`
	Tag                 *Tag                `json:"tag"`
	TagMap              *TagMap             `json:"tagMap"`
	UserMark            *UserMark           `json:"userMark"`
	UserMarkBlockRange  *UserMarkBlockRange `json:"userMarkBlockRange"`
}

// MakeModelSlice converts a slice of pointers of model-implementing structs to []model
func MakeModelSlice(arg interface{}) ([]Model, error) {
	slice := reflect.ValueOf(arg)

	if slice.Kind() != reflect.Kind(reflect.Slice) {
		return nil, fmt.Errorf("Can't create []model out of %T", arg)
	}

	c := slice.Len()
	result := make([]Model, c)
	for i := 0; i < c; i++ {
		result[i] = slice.Index(i).Interface().(Model)
	}
	return result, nil
}

// MakeModelCopy copies the content of the given Model (pointer to a
// model-implementing struct) to a new Model
func MakeModelCopy(mdl Model) Model {
	var mdlCopy Model
	switch mdl.(type) {
	case *BlockRange:
		mdlCopy = &BlockRange{}
	case *Bookmark:
		mdlCopy = &Bookmark{}
	case *Location:
		mdlCopy = &Location{}
	case *Note:
		mdlCopy = &Note{}
	case *Tag:
		mdlCopy = &Tag{}
	case *TagMap:
		mdlCopy = &TagMap{}
	case *UserMark:
		mdlCopy = &UserMark{}
	case *UserMarkBlockRange:
		umbr := mdl.(*UserMarkBlockRange)

		umCopy := &UserMark{}
		copier.Copy(umCopy, umbr.UserMark)

		// Copier doesn't deep copy slices, so doing it manually
		brSliceCopy := make([]*BlockRange, len(umbr.BlockRanges))
		for i, br := range umbr.BlockRanges {
			if br != nil {
				brSliceCopy[i] = &BlockRange{}
				copier.Copy(brSliceCopy[i], umbr.BlockRanges[i])
			}
		}
		return &UserMarkBlockRange{
			UserMark:    umCopy,
			BlockRanges: brSliceCopy,
		}
	default:
		panic(fmt.Sprintf("Type %T is not supported for copying", mdl))
	}

	copier.Copy(mdlCopy, mdl)

	return mdlCopy
}

// PrettyPrint prints the given fields of a Model as a table. If the field
// is empty, its omitted.
func PrettyPrint(m Model, fields []string) string {
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)

Loop:
	for _, fieldName := range fields {
		field := reflect.ValueOf(m).Elem().FieldByName(fieldName)
		if !field.IsValid() {
			panic(fmt.Sprintf("Given struct does not contain field %s", fieldName))
		}
		switch field.Interface().(type) {
		case string:
			fmt.Fprintf(w, "\n%s:\t%s", fieldName, strings.ReplaceAll(wordwrap.WrapString(field.String(), 70), "\n", "\n\t"))
		case sql.NullString:
			if field.Field(1).Bool() == false {
				continue Loop
			}
			fmt.Fprintf(w, "\n%s:\t%s", fieldName, strings.ReplaceAll(wordwrap.WrapString(field.Field(0).String(), 70), "\n", "\n\t"))
		case int:
			fmt.Fprintf(w, "\n%s:\t%d", fieldName, field.Int())
		case sql.NullInt32:
			if field.Field(1).Bool() == false {
				continue Loop
			}
			fmt.Fprintf(w, "\n%s:\t%d", fieldName, field.Field(0).Int())
		default:
			panic(fmt.Sprintf("Unsupported type for field %s", fieldName))
		}
	}
	w.Flush()
	return buf.String()
}

// sortByUniqueKey sorts the given pointer to a slice of Model by UniqueKey,
// removes unnecessary nil-entries (except at position 0),
// and also updates the IDs accordingly. It tracks these changes
// by a map, for which the key represents the old ID,
// and value represents the new ID.
func sortByUniqueKey(slice interface{}) map[int]int {
	changes := map[int]int{}

	if reflect.TypeOf(slice).Kind() != reflect.Ptr || reflect.TypeOf(slice).Elem().Kind() != reflect.Slice {
		panic("Only pointer to slice is supported")
	}

	s := reflect.ValueOf(slice).Elem()

	// Sort by UniqueKey
	sort.Slice(s.Interface(), func(i, j int) bool {
		// Nil is always smaller than every other value
		jVal := s.Index(j)
		if jVal.IsNil() {
			return false
		}
		iVal := s.Index(i)
		if iVal.IsNil() {
			return true
		}

		iUQ := s.Index(i).MethodByName("UniqueKey").Call(nil)
		jUQ := s.Index(j).MethodByName("UniqueKey").Call(nil)
		return iUQ[0].String() < jUQ[0].String()
	})

	// If there are more than one nil values, remove all except one
	// (all nil values are located at the beginning)
	nilCount := 0
	for i := 0; i < s.Len(); i++ {
		if !s.Index(i).IsNil() {
			continue
		}
		nilCount++
	}
	if nilCount > 1 {
		s.Set(s.Slice(nilCount-1, s.Len()))
	}

	// Update IDs to their index
	for i := 0; i < s.Len(); i++ {
		elem := s.Index(i)
		if elem.IsNil() {
			continue
		}
		oldID := int(elem.MethodByName("ID").Call(nil)[0].Int())
		if oldID != i {
			changes[oldID] = i
			elem.MethodByName("SetID").Call([]reflect.Value{reflect.ValueOf(i)})
		}
	}

	return changes
}

// UpdateIDs updates a given ID (named by IDName) on the slice of *model.Model
// according to the given map, for which the key represents the old ID,
// and value represents the new ID.
func UpdateIDs(mdl interface{}, IDName string, changes map[int]int) {
	switch reflect.TypeOf(mdl).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(mdl)
		for i := 0; i < s.Len(); i++ {
			elem := s.Index(i)
			if elem.IsNil() {
				continue
			}

			field := elem.Elem().FieldByName(IDName)
			if !field.IsValid() {
				panic(fmt.Sprintf("Given struct does not contain field %s", IDName))
			}

			switch t := field.Interface().(type) {
			case int:
				if new, ok := changes[int(field.Int())]; ok {
					field.SetInt(int64(new))
				}
			case sql.NullInt32:
				val := field.Field(0)
				if new, ok := changes[int(val.Int())]; ok {
					val.SetInt(int64(new))
				}
			default:
				panic(fmt.Sprintf("Type %T of field %s is not supported!", t, IDName))
			}
		}
	default:
		panic("Only slices are supported!")
	}
}
