package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserMark_SetID(t *testing.T) {
	m1 := &UserMark{UserMarkID: 1}
	m1.SetID(10)
	assert.Equal(t, 10, m1.UserMarkID)

	m2 := UserMark{UserMarkID: 2}
	m2.SetID(20)
	assert.Equal(t, 20, m2.UserMarkID)
}

func TestUserMark_Equals(t *testing.T) {
	m1 := &UserMark{
		UserMarkID:   1,
		ColorIndex:   1,
		LocationID:   1,
		StyleIndex:   1,
		UserMarkGUID: "FIRST",
		Version:      1,
	}
	m1_1 := &UserMark{
		UserMarkID:   1000,
		ColorIndex:   1,
		LocationID:   1,
		StyleIndex:   1,
		UserMarkGUID: "FIRSTT",
		Version:      1,
	}

	m2 := &UserMark{
		UserMarkID:   1,
		ColorIndex:   5,
		LocationID:   1,
		StyleIndex:   1,
		UserMarkGUID: "FIRST",
		Version:      1,
	}
	assert.True(t, m1.Equals(m1_1))
	assert.False(t, m1.Equals(m2))
}