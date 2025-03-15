package userdb

import (
	"database/sql"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type UserDB struct {
	db  *sql.DB
	mux sync.Mutex
}

func NewUserStorage(dbPath string) (*UserDB, error) {
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			return nil, err
		}

		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS users (
				user_id INTEGER PRIMARY KEY,
				username TEXT,
				first_seen DATETIME,
				last_seen DATETIME,
				last_message TEXT,
				download_count INTEGER DEFAULT 1
			)
		`)
		if err != nil {
			return nil, err
		}
		return &UserDB{db: db}, nil
	} else if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &UserDB{db: db}, nil
}

func (us *UserDB) Add(userID int, username, message string) error {
	us.mux.Lock()
	defer us.mux.Unlock()

	now := time.Now().Format(time.RFC3339)
	_, err := us.db.Exec(`
		INSERT INTO users (user_id, username, first_seen, last_seen, last_message)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			username = excluded.username,
			last_seen = excluded.last_seen,
			last_message = excluded.last_message,
			download_count = users.download_count + 1
	`, userID, username, now, now, message)
	return err
}
