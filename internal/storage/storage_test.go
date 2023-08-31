package storage

import (
	"database/sql"
	"github.com/loveletter4you/user-segmentation-service/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

var TEST_DB_URL = "postgres://test:test@localhost:5432/test?sslmode=disable"

func TestCreateStorage(t *testing.T) {
	storage := NewStorage()
	var err error
	storage.db, err = sql.Open("postgres", TEST_DB_URL)
	if err != nil {
		t.Fatal(err)
		return
	}

	user := model.User{}

	err = storage.User().CreateUser(nil, &user)

	assert.NoError(t, err)

	users, err := storage.User().GetUsers(nil)

	assert.NoError(t, err)
	assert.Equal(t, user.Id, users[0].Id)

	segment := model.Segment{Slug: "AVITO_VOICE_MESSAGES"}
	err = storage.Segment().CreateSegment(nil, &segment)
	assert.NoError(t, err)

	segments, err := storage.Segment().GetSegments(nil)

	assert.NoError(t, err)
	assert.Equal(t, segment.Id, segments[0].Id)
	assert.Equal(t, segment.Slug, segments[0].Slug)
	assert.Equal(t, 1, len(segments))

	userSegment, err := storage.Segment().CreateUserSegment(nil, user.Id, segment.Slug, 0)

	assert.NoError(t, err)

	userSegments, err := storage.Segment().GetUserSegments(nil, user.Id)

	assert.NoError(t, err)
	assert.Equal(t, userSegment.Segment.Slug, userSegments[0].Segment.Slug)
}
