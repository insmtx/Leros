// Package seqtracker tracks NATS JetStream message sequence numbers in SQLite
// for crash recovery. It records each received message and its processing status,
// enabling the consumer to replay missed messages on restart.
package seqtracker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Status represents the processing state of a tracked message.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// SeqTracker tracks NATS JetStream message processing status.
type SeqTracker interface {
	// TrackReceived records a newly received message as "pending".
	// Idempotent: uses INSERT OR IGNORE so re-delivery doesn't overwrite.
	TrackReceived(ctx context.Context, topic string, seq uint64, sessionID, messageID, taskID, runID string) error

	// MarkProcessing transitions a message to "processing".
	MarkProcessing(ctx context.Context, topic string, seq uint64) error

	// MarkCompleted transitions a message to "completed".
	MarkCompleted(ctx context.Context, topic string, seq uint64) error

	// MarkFailed transitions a message to "failed" with an error message.
	MarkFailed(ctx context.Context, topic string, seq uint64, errMsg string) error

	// GetLastCompletedSeq returns the highest seq with status=completed for the topic.
	// Returns 0 if no completed records exist.
	GetLastCompletedSeq(ctx context.Context, topic string) (uint64, error)

	// IsDuplicate returns true if the seq has already been completed for this topic.
	IsDuplicate(ctx context.Context, topic string, seq uint64) (bool, error)

	// Close closes the database.
	Close() error
}

// NewSQLiteTracker opens or creates a SQLite database at the given path.
func NewSQLiteTracker(dbPath string) (SeqTracker, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open seq tracker db %s: %w", dbPath, err)
	}
	db.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate seq tracker: %w", err)
	}

	return &sqliteTracker{db: db}, nil
}

type sqliteTracker struct {
	db *sql.DB
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS task_seq (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			topic      TEXT NOT NULL,
			seq        INTEGER NOT NULL,
			session_id TEXT NOT NULL DEFAULT '',
			status     TEXT NOT NULL DEFAULT 'pending',
			message_id TEXT NOT NULL DEFAULT '',
			task_id    TEXT NOT NULL DEFAULT '',
			run_id     TEXT NOT NULL DEFAULT '',
			error_msg  TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_task_seq_unique ON task_seq(topic, seq);
		CREATE INDEX IF NOT EXISTS idx_task_seq_status ON task_seq(status);
	`)
	return err
}

func (s *sqliteTracker) TrackReceived(_ context.Context, topic string, seq uint64, sessionID, messageID, taskID, runID string) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO task_seq (topic, seq, session_id, status, message_id, task_id, run_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		topic, seq, sessionID, string(StatusPending), messageID, taskID, runID, now, now,
	)
	return err
}

func (s *sqliteTracker) MarkProcessing(_ context.Context, topic string, seq uint64) error {
	return s.updateStatus(topic, seq, StatusProcessing, "")
}

func (s *sqliteTracker) MarkCompleted(_ context.Context, topic string, seq uint64) error {
	return s.updateStatus(topic, seq, StatusCompleted, "")
}

func (s *sqliteTracker) MarkFailed(_ context.Context, topic string, seq uint64, errMsg string) error {
	return s.updateStatus(topic, seq, StatusFailed, errMsg)
}

func (s *sqliteTracker) GetLastCompletedSeq(_ context.Context, topic string) (uint64, error) {
	var lastSeq sql.NullInt64
	err := s.db.QueryRow(
		`SELECT MAX(seq) FROM task_seq WHERE topic = ? AND status = ?`,
		topic, string(StatusCompleted),
	).Scan(&lastSeq)
	if err != nil {
		return 0, err
	}
	if !lastSeq.Valid {
		return 0, nil
	}
	return uint64(lastSeq.Int64), nil
}

func (s *sqliteTracker) IsDuplicate(_ context.Context, topic string, seq uint64) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM task_seq WHERE topic = ? AND seq = ? AND status = ?`,
		topic, seq, string(StatusCompleted),
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *sqliteTracker) Close() error {
	return s.db.Close()
}

func (s *sqliteTracker) updateStatus(topic string, seq uint64, status Status, errMsg string) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`UPDATE task_seq SET status = ?, error_msg = ?, updated_at = ? WHERE topic = ? AND seq = ?`,
		string(status), errMsg, now, topic, seq,
	)
	return err
}
