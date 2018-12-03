package model

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "fmt"
  "github.com/buchanae/mailer/imap"
)

const MaxBodyBytes = 10000000

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
    return nil, fmt.Errorf("opening database connection: %s", err)
	}

  _, err = db.Exec(schemaSql)
	if err != nil {
    return nil, fmt.Errorf("creating database schema: %s", err)
	}

  _, err = db.Exec(startupSql)
	if err != nil {
    return nil, fmt.Errorf("configuring database connection: %s", err)
	}
  return &DB{db: db}, nil
}

type DB struct {
  db *sql.DB
}

func (db *DB) Close() error {
  return db.db.Close()
}

func (db *DB) CreateMailbox(name string) error {
  _, err := db.db.Exec("insert into mailbox(name) values(?)", name)
  return err
}

func (db *DB) Message(id int) (*Message, error) {
  msg := &Message{Headers: Headers{}}
  var seen, answered, flagged, deleted, draft, recent bool

  q := `select 
    id,
    size,
    created,
    seen,
    answered,
    flagged,
    deleted,
    draft,
    recent,
    text_path
  from message where id = ?`

  row := db.db.QueryRow(q, id)
  err := row.Scan(
    &msg.ID,
    &msg.Size,
    &msg.Created,
    &seen,
    &answered,
    &flagged,
    &deleted,
    &draft,
    &recent,
    &msg.TextPath,
  )
  if err != nil {
    return nil, fmt.Errorf("loading message from database: %v", err)
  }

  if seen {
    msg.SetFlag(imap.Seen)
  }
  if answered {
    msg.SetFlag(imap.Answered)
  }
  if flagged {
    msg.SetFlag(imap.Flagged)
  }
  if deleted {
    msg.SetFlag(imap.Deleted)
  }
  if draft {
    msg.SetFlag(imap.Draft)
  }
  if recent {
    msg.SetFlag(imap.Recent)
  }

  rows, err := db.db.Query(`select key, value from header where message_id = ?`, msg.ID)
  if err != nil {
    return nil, fmt.Errorf("loading message headers from database: %v", err)
  }
  defer rows.Close()

  for rows.Next() {
    var key, value string
    err := rows.Scan(&key, &value)
    if err != nil {
      return nil, fmt.Errorf("loading message headers from database: %v", err)
    }
    msg.Headers[key] = append(msg.Headers[key], value)
  }
  if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("loading message headers from database: %v", err)
  }

  return msg, nil
}

func (db *DB) ListMailboxes() ([]*Mailbox, error) {

  var boxes []*Mailbox

  rows, err := db.db.Query("select id, name from mailbox")
  if err != nil {
    return nil, fmt.Errorf("loading mailboxes from database: %v", err)
  }
  defer rows.Close()

  for rows.Next() {
    box := &Mailbox{}
    err := rows.Scan(&box.ID, &box.Name)
    if err != nil {
      return nil, fmt.Errorf("loading mailboxes from database: %v", err)
    }
    boxes = append(boxes, box)
  }

  if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("loading message headers from database: %v", err)
  }

  return boxes, nil
}

func (db *DB) MailboxByName(name string) (*Mailbox, error) {

  box := &Mailbox{Name: name}
  q := "select id from mailbox where name = ?"
  row := db.db.QueryRow(q, name)
  err := row.Scan(&box.ID)
  if err == sql.ErrNoRows {
    return nil, fmt.Errorf("no mailbox named %q", name)
  }
  if err != nil {
    return nil, fmt.Errorf("finding mailbox by name: %v", err)
  }
  return box, nil
}

func (db *DB) CreateMail(box *Mailbox, msg *Message) error {
  // TODO probably want to return the created mail ID
  return db.withTx(func() error {

    var seen, answered, flagged, deleted, draft, recent bool
    for _, flag := range msg.Flags {
      switch flag {
      case imap.Seen:
        seen = true
      case imap.Answered:
        answered = true
      case imap.Flagged:
        flagged = true
      case imap.Deleted:
        deleted = true
      case imap.Draft:
        draft = true
      case imap.Recent:
        recent = true
      }
    }

    res, err := db.db.Exec(
    `insert into message(
      mailbox_id,
      size,
      created,
      seen,
      answered,
      flagged,
      deleted,
      draft,
      recent,
      content
    ) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      box.ID,
      msg.Size,
      msg.Created,
      seen,
      answered,
      flagged,
      deleted,
      draft,
      recent,
      msg.TextPath,
    )
    if err != nil {
      return fmt.Errorf("inserting mail into database: %v", err)
    }

    msgID, err := res.LastInsertId()
    if err != nil {
      return fmt.Errorf("getting inserted mail ID: %v", err)
    }

    for key, values := range msg.Headers {
      for _, value := range values {
        _, err = db.db.Exec("insert into header(message_id, key, value) values (?, ?, ?)",
          msgID, key, value)
        if err != nil {
          return fmt.Errorf("inserting header into database: %v", err)
        }
      }
    }
    return nil
  })
}

func (db *DB) withTx(f func() error) error {
  tx, err := db.db.Begin()
  if err != nil {
    return fmt.Errorf("beginning transaction: %v", err)
  }
  err = f()
  if err != nil {
    if rollbackErr := tx.Rollback(); rollbackErr != nil {
      return fmt.Errorf("%v\nfailed to roll back transaction: %v", err, rollbackErr)
    }
    return fmt.Errorf("rolled back transaction: %v", err)
  }
  commitErr := tx.Commit()
  if commitErr != nil {
    return fmt.Errorf("failed to commit transaction: %v", commitErr)
  }
  return nil
}
