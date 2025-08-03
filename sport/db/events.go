package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shanehowearth/entain-master/sport/proto/sporting"
)

// EventsRepo provides repository access to events.
type EventsRepo interface {
	// Init will initialise our events repository.
	Init() error

	// Get will return a event.
	Get(eventId int64) (*sporting.Event, error)

	// List will return a list of events.
	List(filter *sporting.ListEventsRequestFilter) ([]*sporting.Event, error)
}

type eventsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewEventsRepo creates a new events repository.
func NewEventsRepo(db *sql.DB) EventsRepo {
	return &eventsRepo{db: db}
}

// Init prepares the event repository dummy data.
func (e *eventsRepo) Init() error {
	var err error

	e.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy events.
		err = e.seed()
	})

	return err
}

func (e *eventsRepo) Get(eventId int64) (*sporting.Event, error) {
	var (
		err   error
		query string
	)

	query = getEventQueries()[eventsList]

	query += " WHERE id = ?"
	rows, err := e.db.Query(query, eventId)
	if err != nil {
		return nil, err
	}

	events, err := e.scanEvents(rows)
	if err != nil {
		return nil, err
	}

	return events[0], nil
}

func (e *eventsRepo) List(filter *sporting.ListEventsRequestFilter) ([]*sporting.Event, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getEventQueries()[eventsList]

	query, args = e.applyFilter(query, filter)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return e.scanEvents(rows)
}

func (e *eventsRepo) applyFilter(query string, filter *sporting.ListEventsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}

	// Visible filter.
	if filter.Visible != nil {
		clauses = append(clauses, "visible = ?")
		visible := 0
		if *filter.Visible {
			visible = 1
		}
		args = append(args, visible)
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	// Order filter.
	// This allows the user to sort the rows by advertised start time,
	// Ascending, or Descending.
	// ORDER BY date_column DESC;
	// Note: This clause MUST be after the WHERE, and any HAVING, clauses, but
	// before any LIMIT or OFFSET clauses.
	// Other order by clauses for different fields can be added here too.
	// Finally: advertisedStartTime is hard coded here, but there's no reason
	// that it cannot be another variable passed in by the user.
	if filter.SortDirection != nil {
		direction := ""
		switch *filter.SortDirection {
		case 1:
			direction = " ORDER BY advertised_start_time ASC"
		case 2:
			direction = " ORDER BY advertised_start_time DESC"
		}
		query += direction
	}

	return query, args
}

// testableTimeNow - this can be overriden by a test in the same package,
// allowing deterministic testing without having to manipulate the event data.
var testableTimeNow = time.Now

func (m *eventsRepo) scanEvents(
	rows *sql.Rows,
) ([]*sporting.Event, error) {
	var events []*sporting.Event

	for rows.Next() {
		var event sporting.Event
		var advertisedStart time.Time

		if err := rows.Scan(&event.Id, &event.MeetingId, &event.Name, &event.Number, &event.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		event.AdvertisedStartTime = ts

		// Set event status as open or closed.
		// A event is closed if its advertised starttime has passed.
		// Note: the cut off time is hardcoded to Now, but there's no reason
		// that it couldn't be passed in as a variable if desired.
		status := "OPEN"
		if advertisedStart.Before(testableTimeNow()) {
			status = "CLOSED"
		}

		event.Status = status

		events = append(events, &event)
	}

	return events, nil
}
