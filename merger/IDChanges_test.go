package merger

import (
	"database/sql"
	"testing"

	"github.com/AndreasSko/go-jwlm/model"
	"github.com/stretchr/testify/assert"
)

func TestUpdateIDs_Notes(t *testing.T) {
	left := []*model.Note{
		nil,
		{
			NoteID:     1,
			LocationID: sql.NullInt32{1, true},
			BlockType:  1,
		},
		{
			NoteID:     2,
			LocationID: sql.NullInt32{2, true},
		},
		nil,
		{
			NoteID:     3,
			LocationID: sql.NullInt32{1, true},
		},
	}
	right := []*model.Note{
		nil,
		nil,
		{},
		{
			NoteID:     1,
			LocationID: sql.NullInt32{1, true},
		},
		{
			NoteID:     2,
			LocationID: sql.NullInt32{5, true},
		},
	}
	changes := IDChanges{
		Left: map[int]int{
			1: 5,
		},
		Right: map[int]int{
			5: 3,
		},
	}

	expectedLeft := []*model.Note{
		nil,
		{
			NoteID:     1,
			LocationID: sql.NullInt32{5, true},
			BlockType:  1,
		},
		{
			NoteID:     2,
			LocationID: sql.NullInt32{2, true},
		},
		nil,
		{
			NoteID:     3,
			LocationID: sql.NullInt32{5, true},
		},
	}
	expectedRight := []*model.Note{
		nil,
		nil,
		{},
		{
			NoteID:     1,
			LocationID: sql.NullInt32{1, true},
		},
		{
			NoteID:     2,
			LocationID: sql.NullInt32{3, true},
		},
	}

	UpdateIDs(left, right, "LocationID", changes)
	assert.Equal(t, expectedLeft, left)
	assert.Equal(t, expectedRight, right)
}

func TestUpdateIDs_Bookmarks(t *testing.T) {
	left := []*model.Bookmark{
		nil,
		{
			BookmarkID: 1,
			LocationID: 1,
		},
		{
			BookmarkID: 2,
		},
	}
	right := []*model.Bookmark{
		nil,
		{},
		{
			BookmarkID: 2,
			LocationID: 5,
		},
		{
			BookmarkID: 3,
			LocationID: 1,
		},
	}
	changes := IDChanges{
		Left: map[int]int{
			1: 5,
		},
		Right: map[int]int{
			1: 2,
		},
	}

	expectedLeft := []*model.Bookmark{
		nil,
		{
			BookmarkID: 1,
			LocationID: 5,
		},
		{
			BookmarkID: 2,
		},
	}
	expectedRight := []*model.Bookmark{
		nil,
		{},
		{
			BookmarkID: 2,
			LocationID: 5,
		},
		{
			BookmarkID: 3,
			LocationID: 2,
		},
	}

	UpdateIDs(left, right, "LocationID", changes)
	assert.Equal(t, expectedLeft, left)
	assert.Equal(t, expectedRight, right)

	assert.PanicsWithValue(t, "Given struct does not contain field WrongField", func() {
		UpdateIDs(left, right, "WrongField", changes)
	})
	assert.PanicsWithValue(t, "Type string of field Title is not supported!", func() {
		UpdateIDs(left, right, "Title", changes)
	})
	assert.PanicsWithValue(t, "Only slices are supported!", func() {
		UpdateIDs(model.Bookmark{}, model.Bookmark{}, "LocationID", changes)
	})
}
