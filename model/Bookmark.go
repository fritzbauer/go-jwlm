package model

import (
	"database/sql"
	"strconv"
	"strings"
)

// Bookmark represents the Bookmark table inside the JW Library database
type Bookmark struct {
	BookmarkID            int
	LocationID            int
	PublicationLocationID int
	Slot                  int
	Title                 string
	Snippet               sql.NullString
	BlockType             int
	BlockIdentifier       sql.NullInt32
}

// ID returns the ID of the entry
func (m *Bookmark) ID() int {
	return m.BookmarkID
}

// SetID sets the ID of the entry
func (m *Bookmark) SetID(id int) {
	m.BookmarkID = id
}

// UniqueKey returns the key that makes this Bookmark unique,
// so it can be used as a key in a map.
func (m *Bookmark) UniqueKey() string {
	var sb strings.Builder
	sb.Grow(6)
	sb.WriteString(strconv.FormatInt(int64(m.PublicationLocationID), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.Slot), 10))
	return sb.String()
}

// Equals checks if the Bookmark is equal to the given one. The
// check won't include the BookmarkID
func (m *Bookmark) Equals(m2 Model) bool {
	if m2, ok := m2.(*Bookmark); ok {
		return m.LocationID == m2.LocationID &&
			m.PublicationLocationID == m2.PublicationLocationID &&
			m.Slot == m2.Slot &&
			m.Title == m2.Title &&
			m.Snippet == m2.Snippet &&
			m.BlockType == m2.BlockType &&
			m.BlockIdentifier == m2.BlockIdentifier
	}

	return false
}

func (m *Bookmark) tableName() string {
	return "Bookmark"
}

func (m *Bookmark) idName() string {
	return "BookmarkId"
}

func (m *Bookmark) scanRow(rows *sql.Rows) (Model, error) {
	err := rows.Scan(&m.BookmarkID, &m.LocationID, &m.PublicationLocationID, &m.Slot, &m.Title,
		&m.Snippet, &m.BlockType, &m.BlockIdentifier)
	return m, err
}

// MakeSlice converts a slice of the generice interface model
func (Bookmark) MakeSlice(mdl []Model) []*Bookmark {
	result := make([]*Bookmark, len(mdl))
	for i := range mdl {
		if mdl[i] != nil {
			result[i] = mdl[i].(*Bookmark)
		}
	}
	return result
}
