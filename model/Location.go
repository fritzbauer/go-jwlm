package model

import (
	"database/sql"
	"strconv"
	"strings"
)

// Location represents the Location table inside the JW Library database
type Location struct {
	LocationID     int
	BookNumber     sql.NullInt32
	ChapterNumber  sql.NullInt32
	DocumentID     sql.NullInt32
	Track          sql.NullInt32
	IssueTagNumber int
	KeySymbol      sql.NullString
	MepsLanguage   int
	LocationType   int
	Title          sql.NullString
}

// ID returns the ID of the entry
func (m *Location) ID() int {
	return m.LocationID
}

// SetID sets the ID of the entry
func (m *Location) SetID(id int) {
	m.LocationID = id
}

// UniqueKey returns the key that makes this Location unique,
// so it can be used as a key in a map.
func (m *Location) UniqueKey() string {
	var sb strings.Builder
	sb.Grow(35)
	sb.WriteString(strconv.FormatInt(int64(m.BookNumber.Int32), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.ChapterNumber.Int32), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.DocumentID.Int32), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.Track.Int32), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.IssueTagNumber), 10))
	sb.WriteString("_")
	sb.WriteString(m.KeySymbol.String)
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.MepsLanguage), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.LocationType), 10))
	return sb.String()
}

// Equals checks if the Location is equal to the given one.
func (m *Location) Equals(m2 Model) bool {
	return false
}

func (m *Location) tableName() string {
	return "Location"
}

func (m *Location) idName() string {
	return "LocationId"
}

func (m *Location) scanRow(rows *sql.Rows) (Model, error) {
	err := rows.Scan(&m.LocationID, &m.BookNumber, &m.ChapterNumber, &m.DocumentID, &m.Track,
		&m.IssueTagNumber, &m.KeySymbol, &m.MepsLanguage, &m.LocationType, &m.Title)
	return m, err
}

// MakeSlice converts a slice of the generice interface model
func (Location) MakeSlice(mdl []Model) []*Location {
	result := make([]*Location, len(mdl))
	for i := range mdl {
		if mdl[i] != nil {
			result[i] = mdl[i].(*Location)
		}
	}
	return result
}