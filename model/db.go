package model

import (
  "io"
  "time"
  "os"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "fmt"
  "github.com/buchanae/mailer/imap"
  "path/filepath"
)

const MaxBodyBytes = 10000000

func ensureDir(path string) error {
  // Check that the data directory exists.
  s, err := os.Stat(path)
  if os.IsNotExist(err) {
    err := os.Mkdir(path, 0700)
    if err != nil {
      return fmt.Errorf("creating data directory: %v", err)
    }
    return nil
  } else if err != nil {
    return fmt.Errorf("checking for data directory: %v", err)
  }

  if !s.IsDir() {
    return fmt.Errorf("%q is a file, but mailer needs to put a directory here", path)
  }
  return nil
}

func Open(path string) (*DB, error) {
  err := ensureDir(path)
  if err != nil {
    return nil, err
  }

  err = ensureDir(filepath.Join(path, "messages"))
  if err != nil {
    return nil, err
  }

  // Open the sqlite database.
	db, err := sql.Open("sqlite3", filepath.Join(path, "mailer.db"))
	if err != nil {
    return nil, fmt.Errorf("opening database connection: %s", err)
	}

  // Set up the schema.
  _, err = db.Exec(schemaSql)
	if err != nil {
    return nil, fmt.Errorf("creating database schema: %s", err)
	}

  // Configure the database engine.
  _, err = db.Exec(startupSql)
	if err != nil {
    return nil, fmt.Errorf("configuring database connection: %s", err)
	}

  return &DB{path: path, db: db}, nil
}

type DB struct {
  path string
  db *sql.DB
}

func (db *DB) Close() error {
  return db.db.Close()
}

func (db *DB) CreateMailbox(name string) error {
  _, err := db.db.Exec("insert into mailbox(name) values(?)", name)
  return err
}

func (db *DB) RenameMailbox(from, to string) error {
  _, err := db.db.Exec("update mailbox set name = ? where name = ?", to, from)
  return err
}

func (db *DB) DeleteMailbox(name string) error {
  _, err := db.db.Exec("delete from mailbox where name = ?", name)
  return err
}

func (db *DB) SetFlags(id int, flags ...imap.Flag) error {
  return db.withTx(func(tx *sql.Tx) error {
    for _, flag := range flags {

      _, err := tx.Exec(
        "insert or ignore into flag (message_id, value) values (?, ?)",
        id, flag)

      if err != nil {
        return err
      }
    }
    return nil
  })
}

func (db *DB) MessageRange(offset, limit int) ([]*Message, error) {
  return nil, fmt.Errorf("database error: message range not implemented")
}

func (db *DB) Message(id int) (*Message, error) {
  msg := &Message{Headers: Headers{}}

  q := `select 
    id,
    size,
    created,
    text_path
  from message where id = ?`

  row := db.db.QueryRow(q, id)
  err := row.Scan(
    &msg.ID,
    &msg.Size,
    &msg.Created,
    &msg.TextPath,
  )
  if err != nil {
    return nil, fmt.Errorf("loading message from database: %v", err)
  }

  err = db.loadFlags(msg)
  if err != nil {
    return nil, fmt.Errorf("loading message from database: %v", err)
  }

  err = db.loadHeaders(msg)
  if err != nil {
    return nil, fmt.Errorf("loading message from database: %v", err)
  }

  return msg, nil
}

func (db *DB) loadFlags(msg *Message) error {
  rows, err := db.db.Query(`select value from flag where message_id = ?`, msg.ID)
  if err != nil {
    return fmt.Errorf("loading message flags: %v", err)
  }
  defer rows.Close()

  for rows.Next() {
    var value string
    err := rows.Scan(&value)
    if err != nil {
      return fmt.Errorf("loading message flags: %v", err)
    }
    flag := imap.LookupFlag(value)
    msg.Flags = append(msg.Flags, flag)
  }

  if err := rows.Err(); err != nil {
    return fmt.Errorf("loading message flags: %v", err)
  }
  return nil
}

func (db *DB) loadHeaders(msg *Message) error {
  rows, err := db.db.Query(`select key, value from header where message_id = ?`, msg.ID)
  if err != nil {
    return fmt.Errorf("loading message headers: %v", err)
  }
  defer rows.Close()

  for rows.Next() {
    var key, value string
    err := rows.Scan(&key, &value)
    if err != nil {
      return fmt.Errorf("loading message headers: %v", err)
    }
    msg.Headers[key] = append(msg.Headers[key], value)
  }

  if err := rows.Err(); err != nil {
    return fmt.Errorf("loading message headers: %v", err)
  }
  return nil
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

var errByteLimitReached = fmt.Errorf("max byte limit reached")

// maxReader limits the number of bytes read from the underlying reader "R",
// and returns an errByteLimitReached if the limit is reached.
type maxReader struct {
  R io.Reader // underlying reader
  N int // max bytes remaining
}

func (m *maxReader) Read(p []byte) (int, error) {
  if len(p) > m.N {
    return 0, errByteLimitReached
  }
  n, err := m.R.Read(p)
  m.N -= n
  return n, err
}

func (db *DB) CreateMail(box *Mailbox, msg *Message, text io.Reader) error {
  return db.withTx(func(tx *sql.Tx) error {
    created := time.Now()

    // Insert an empty row in order to get/reserve the next message ID.
    res, err := tx.Exec(
      `insert into message(
        mailbox_id,
        size,
        created,
        text_path
      ) values (?, ?, ?, ?)`,
      box.ID, 0, created, "")
    if err != nil {
      return fmt.Errorf("inserting mail into database: %v", err)
    }

    msgID, err := res.LastInsertId()
    if err != nil {
      return fmt.Errorf("getting inserted mail ID: %v", err)
    }

    for key, values := range msg.Headers {
      for _, value := range values {

        _, err := tx.Exec(
          "insert into header(message_id, key, value) values (?, ?, ?)",
          msgID, key, value)

        if err != nil {
          return fmt.Errorf("inserting header into database: %v", err)
        }
      }
    }

    for _, flag := range msg.Flags {

      _, err := tx.Exec(
        "insert into flag(message_id, value) values (?, ?)",
        msgID, flag)

      if err != nil {
        return fmt.Errorf("database error: inserting flag: %v", err)
      }
    }

    // Write the message body to a file.

    // Split the files into groups of 1000.
    msgDir := filepath.Join(db.path, "messages", fmt.Sprint(msgID % 1000))
    err = ensureDir(msgDir)
    if err != nil {
      return fmt.Errorf("creating message body file: %v", err)
    }

    msgPath := filepath.Join(msgDir, fmt.Sprint(msgID))
    fh, err := os.Create(msgPath)
    if err != nil {
      return fmt.Errorf("creating message body file: %v", err)
    }
    defer fh.Close()

    // Limit the size of the message body.
    mr := &maxReader{R: text, N: MaxBodyBytes}

    // Copy the data to the file.
    size, err := io.Copy(fh, mr)
    if err != nil {
      os.Remove(msgPath)
      if err == errByteLimitReached {
        return fmt.Errorf("message body is too big. max is %d bytes.", MaxBodyBytes)
      }
      return fmt.Errorf("writing message body file: %v", err)
    }

    // Save some more information in the database: size, path, etc.
    _, err = tx.Exec(`
      update message set size = ?, text_path = ? where id = ?
      `, size, msgPath, msgID)

    if err != nil {
      os.Remove(msgPath)
      return fmt.Errorf("database error: saving message: %v", err)
    }

    msg.ID = msgID
    msg.Size = size
    msg.Created = created
    msg.TextPath = msgPath

    return nil
  })
}

func (db *DB) withTx(f func(*sql.Tx) error) error {
  tx, err := db.db.Begin()
  if err != nil {
    return fmt.Errorf("beginning transaction: %v", err)
  }
  err = f(tx)
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
