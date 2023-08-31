package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/loveletter4you/user-segmentation-service/internal/model"
	"math/rand"
)

type SegmentRepository struct {
	storage *Storage
}

func (sr *SegmentRepository) CreateSegment(tx *sql.Tx, segment *model.Segment) error {
	query := fmt.Sprintf("INSERT INTO segments (slug) VALUES ($$%s$$) RETURNING id", segment.Slug)
	err := sr.storage.DoQueryRow(tx, query).Scan(&segment.Id)
	return err
}

func (sr *SegmentRepository) GetSegments(tx *sql.Tx) ([]*model.Segment, error) {
	segments := make([]*model.Segment, 0)
	query := "SELECT id, slug FROM segments"
	rows, err := sr.storage.DoQuery(tx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		segment := &model.Segment{}
		if err := rows.Scan(&segment.Id, &segment.Slug); err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}
	return segments, err
}

func (sr *SegmentRepository) CreateUserSegment(tx *sql.Tx, userId int, slug string, timeToLive uint) (*model.UserSegment, error) {
	var count int
	query := fmt.Sprintf("SELECT count(user_segment.id) "+
		"FROM user_segment JOIN segments s on user_segment.segment_id = s.id "+
		"WHERE s.slug = '%s' AND user_id = %d AND active_to > now()",
		slug, userId)
	if err := sr.storage.DoQueryRow(tx, query).Scan(&count); err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("user have current active segment")
	}
	if timeToLive == 0 {
		query = fmt.Sprintf("INSERT INTO user_segment (user_id, segment_id) "+
			"SELECT %d, segments.id FROM segments "+
			"WHERE segments.slug = '%s' AND %d in (SELECT id FROM users)"+
			"RETURNING id",
			userId, slug, userId)
	} else {
		query = fmt.Sprintf("INSERT INTO user_segment (user_id, segment_id, active_to) "+
			"SELECT %d, segments.id, now() + interval '%d seconds' FROM segments "+
			"WHERE segments.slug = '%s' AND %d in (SELECT id FROM users)"+
			"RETURNING id",
			userId, timeToLive, slug, userId)
	}
	var userSegment model.UserSegment
	err := sr.storage.DoQueryRow(tx, query).Scan(&userSegment.Id)
	return &userSegment, err
}

func (sr *SegmentRepository) DeleteUserSegment(tx *sql.Tx, userId int, slug string) (*model.UserSegment, error) {
	var segmentId int
	query := fmt.Sprintf("SELECT s.id FROM user_segment JOIN segments s on s.id = user_segment.segment_id "+
		"WHERE (s.slug = '%s' AND user_id = %d AND active_to > now()) GROUP BY s.id",
		slug, userId)
	if err := sr.storage.DoQueryRow(tx, query).Scan(&segmentId); err != nil {
		return nil, errors.New("user have not current active segment")
	}
	query = fmt.Sprintf("UPDATE user_segment SET active_to = now() "+
		"WHERE segment_id = %d AND user_id = %d AND active_to > now() RETURNING id",
		segmentId, userId)
	var userSegment model.UserSegment
	err := sr.storage.DoQueryRow(tx, query).Scan(&userSegment.Id)
	return &userSegment, err
}

func (sr *SegmentRepository) GetUserSegments(tx *sql.Tx, userId int) ([]*model.UserSegment, error) {
	userSegments := make([]*model.UserSegment, 0)
	query := fmt.Sprintf("SELECT segment_id, segments.slug, active_from, active_to "+
		"FROM user_segment JOIN segments ON segment_id = segments.id WHERE user_id = %d "+
		"and active_to > now()", userId)
	rows, err := sr.storage.DoQuery(tx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		userSegment := &model.UserSegment{
			Segment: &model.Segment{},
		}
		if err := rows.Scan(&userSegment.Segment.Id, &userSegment.Segment.Slug,
			&userSegment.ActiveFrom, &userSegment.ActiveTo); err != nil {
			return nil, err
		}
		userSegments = append(userSegments, userSegment)
	}
	return userSegments, nil
}

func (sr *SegmentRepository) GetUserSegmentsMonthYear(tx *sql.Tx, userId, month, year int) ([]*model.UserSegment, error) {
	userSegments := make([]*model.UserSegment, 0)
	dateStr := fmt.Sprintf("%d-%02d", year, month)
	query := fmt.Sprintf("SELECT segment_id, segments.slug, active_from, active_to "+
		"FROM user_segment JOIN segments ON segment_id = segments.id "+
		"where user_id = %d and (to_char(active_from, 'YYYY-MM') = '%s' or to_char(active_to, 'YYYY-MM') = '%s')",
		userId, dateStr, dateStr)
	rows, err := sr.storage.DoQuery(tx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		userSegment := &model.UserSegment{
			Segment: &model.Segment{},
		}
		if err := rows.Scan(&userSegment.Segment.Id, &userSegment.Segment.Slug,
			&userSegment.ActiveFrom, &userSegment.ActiveTo); err != nil {
			return nil, err
		}
		userSegments = append(userSegments, userSegment)
	}
	return userSegments, nil
}

func (sr *SegmentRepository) CreateSegmentAutoInsert(tx *sql.Tx, slug string, percent int, timeToLive uint) (*model.SegmentAutoInsert, error) {
	var count int
	query := fmt.Sprintf("SELECT count(segments_auto_insert.id) "+
		"FROM segments_auto_insert JOIN segments s on segments_auto_insert.segment_id = s.id "+
		"WHERE s.slug = '%s' AND active_to > now()", slug)
	if err := sr.storage.DoQueryRow(tx, query).Scan(&count); err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("current segment is auto inserting")
	}
	chance := float64(percent) / float64(100)
	if timeToLive == 0 {
		query = fmt.Sprintf("INSERT INTO segments_auto_insert (segment_id, chance) "+
			"SELECT segments.id, %.2f FROM segments WHERE segments.slug = '%s' RETURNING id, chance", chance, slug)
	} else {
		query = fmt.Sprintf("INSERT INTO segments_auto_insert (segment_id, chance, active_to) "+
			"SELECT segments.id, %.2f, now() + interval '%d seconds' "+
			"FROM segments WHERE segments.slug = '%s' RETURNING id, chance",
			chance, timeToLive, slug)
	}
	segmentAutoInsert := model.SegmentAutoInsert{
		Segment: &model.Segment{Slug: slug},
	}
	err := sr.storage.DoQueryRow(tx, query).Scan(&segmentAutoInsert.Id, &segmentAutoInsert.Chance)
	return &segmentAutoInsert, err
}

func (sr *SegmentRepository) GetSegmentsAutoInsert(tx *sql.Tx) ([]*model.SegmentAutoInsert, error) {
	segmentsAutoInsert := make([]*model.SegmentAutoInsert, 0)
	query := "SELECT segments_auto_insert.id, s.id, s.slug, chance, active_from, active_to FROM segments_auto_insert " +
		"JOIN segments s on segments_auto_insert.segment_id = s.id WHERE active_to > now()"
	rows, err := sr.storage.DoQuery(tx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		segmentAutoInsert := &model.SegmentAutoInsert{
			Segment: &model.Segment{},
		}
		if err := rows.Scan(&segmentAutoInsert.Id, &segmentAutoInsert.Segment.Id, &segmentAutoInsert.Segment.Slug,
			&segmentAutoInsert.Chance, &segmentAutoInsert.ActiveFrom, &segmentAutoInsert.ActiveTo); err != nil {
			return nil, err
		}
		segmentsAutoInsert = append(segmentsAutoInsert, segmentAutoInsert)
	}
	return segmentsAutoInsert, nil
}

func (sr *SegmentRepository) AutoCreateUserSegments(tx *sql.Tx, users []*model.User, segmentAutoInsert *model.SegmentAutoInsert) {
	for _, user := range users {
		randFloat := rand.Float64()
		if segmentAutoInsert.Chance > randFloat {
			_, err := sr.CreateUserSegment(tx, user.Id, segmentAutoInsert.Segment.Slug, 0)
			if err != nil {
				continue
			}
		}
	}
}

func (sr *SegmentRepository) AutoAddUserToSegments(tx *sql.Tx, user *model.User) error {
	segmentsAutoInsert, err := sr.GetSegmentsAutoInsert(tx)
	if err != nil {
		return err
	}
	for _, segmentAutoInsert := range segmentsAutoInsert {
		randFloat := rand.Float64()
		if segmentAutoInsert.Chance > randFloat {
			_, err = sr.CreateUserSegment(tx, user.Id, segmentAutoInsert.Segment.Slug, 0)
			if err != nil {
				continue
			}
		}
	}
	return nil
}
