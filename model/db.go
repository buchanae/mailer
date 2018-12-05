package model

import (
  "io"
  "io/ioutil"
  "time"
  "os"
  "database/sql"
  "net/mail"
  _ "github.com/mattn/go-sqlite3"
  "fmt"
  "github.com/buchanae/mailer/imap"
  "path/filepath"
)

const MaxBodyBytes = 10000000

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
  _, err = db.Exec(packed)
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

func (db *DB) MessageIDRange(mailbox string, start, end int) ([]*Message, error) {
  var msgs []*Message

  q := `select
    m.row_id,
    m.id,
    m.size,
    m.created,
    m.path
  from message as m
  join mailbox as b
  on m.mailbox_id = b.id
  where b.name = ?
  and m.id >= ?`
  args := []interface{}{mailbox, start}

  if end > start {
    q += ` and m.id <= ?`
    args = append(args, end)
  }

  rows, err := db.db.Query(q, args...)

  if err != nil {
    return nil, fmt.Errorf("loading message range: %v", err)
  }
  defer rows.Close()

  for rows.Next() {
    m := &Message{Headers: Headers{}}
    err := rows.Scan(&m.RowID, &m.ID, &m.Size, &m.Created, &m.Path)
    if err != nil {
      return nil, fmt.Errorf("loading message: %v", err)
    }

    err = db.loadHeaders(m)
    if err != nil {
      return nil, fmt.Errorf("loading message: %v", err)
    }

    err = db.loadFlags(m)
    if err != nil {
      return nil, fmt.Errorf("loading message: %v", err)
    }
    msgs = append(msgs, m)
  }

  if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("loading message range: %v", err)
  }
  return msgs, nil
}

func (db *DB) MessageRange(mailbox string, offset, limit int) ([]*Message, error) {
  var msgs []*Message
  rows, err := db.db.Query(
    `select
      m.row_id,
      m.id,
      m.size,
      m.created,
      m.path
    from message as m
    join mailbox as b
    on m.mailbox_id = b.id
    where b.name = ?
    limit ? offset ?
    `,
    mailbox, limit, offset)

  if err != nil {
    return nil, fmt.Errorf("loading message range: %v", err)
  }
  defer rows.Close()

  for rows.Next() {
    m := &Message{Headers: Headers{}}
    err := rows.Scan(&m.RowID, &m.ID, &m.Size, &m.Created, &m.Path)
    if err != nil {
      return nil, fmt.Errorf("loading message: %v", err)
    }

    err = db.loadHeaders(m)
    if err != nil {
      return nil, fmt.Errorf("loading message: %v", err)
    }

    err = db.loadFlags(m)
    if err != nil {
      return nil, fmt.Errorf("loading message: %v", err)
    }
    msgs = append(msgs, m)
  }

  if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("loading message range: %v", err)
  }
  return msgs, nil
}

func (db *DB) Message(rowID int) (*Message, error) {
  msg := &Message{Headers: Headers{}}

  q := `select 
    row_id,
    id,
    size,
    created,
    path
  from message where row_id = ?`

  row := db.db.QueryRow(q, rowID)
  err := row.Scan(
    &msg.RowID,
    &msg.ID,
    &msg.Size,
    &msg.Created,
    &msg.Path,
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

func (db *DB) messageBodyPath(boxID, msgID int) (string, error) {
  boxDir := filepath.Join(db.path, "messages", fmt.Sprint(boxID))
  // Split the files into groups of 1000.
  msgDir := filepath.Join(boxDir, fmt.Sprint(msgID / 1000))

  err := ensureDir(msgDir)
  if err != nil {
    return "", err
  }

  msgPath := filepath.Join(msgDir, fmt.Sprint(msgID))
  return msgPath, nil
}

func (db *DB) createMessageFile(boxID, msgID int) (*os.File, error) {
  msgPath, err := db.messageBodyPath(boxID, msgID)
  if err != nil {
    return nil, fmt.Errorf("creating message body file: %v", err)
  }

  // TODO probably should ensure the file doesn't already exist for safety.
  fh, err := os.Create(msgPath)
  if err != nil {
    return nil, fmt.Errorf("creating message body file: %v", err)
  }
  return fh, nil
}

func (db *DB) addHeaders(tx *sql.Tx, rowID int, headers Headers) error {
  for key, values := range headers {
    for _, value := range values {

      _, err := tx.Exec(
        "insert into header(message_row_id, key, value) values (?, ?, ?)",
        rowID, key, value)

      if err != nil {
        return fmt.Errorf("inserting header into database: %v", err)
      }
    }
  }
  return nil
}

func (db *DB) CreateMessage(mailbox string, body io.Reader, flags []imap.Flag) (*Message, error) {
  var msg *Message

  dberr := db.withTx(func(tx *sql.Tx) error {

    boxID, msgID, err := db.nextID(tx, mailbox)
    if err != nil {
      return err
    }

    fh, err := db.createMessageFile(boxID, msgID)
    if err != nil {
      return err
    }
    defer fh.Close()
    defer func() {
      if err != nil {
        os.Remove(fh.Name())
      }
    }()

    headers, size, err := saveMessageBody(body, fh)
    if err != nil {
      return err
    }

    msg = &Message{
      ID: int64(msgID),
      Size: size,
      Headers: headers,
      Flags: flags,
      Created: time.Now(),
      Path: fh.Name(),
    }
    err = db.insertMessage(tx, boxID, msg)
    if err != nil {
      return fmt.Errorf("database error: inserting message: %v", err)
    }

    return nil
  })

  return msg, dberr
}

func (db *DB) CopyMessage(msg *Message, to string) (*Message, error) {
  var res *Message

  dberr := db.withTx(func(tx *sql.Tx) error {

    boxID, msgID, err := db.nextID(tx, to)
    if err != nil {
      return err
    }

    path, err := db.messageBodyPath(boxID, msgID)
    if err != nil {
      return err
    }
    os.Link(msg.Path, path)
    defer func() {
      if err != nil {
        os.Remove(path)
      }
    }()

    res = &Message{
      ID: int64(msgID),
      Size: msg.Size,
      Headers: msg.Headers,
      Flags: msg.Flags,
      Created: msg.Created,
      Path: path,
    }
    res.SetFlag(imap.Recent)

    err = db.insertMessage(tx, boxID, res)
    if err != nil {
      return fmt.Errorf("database error: inserting message: %v", err)
    }
    return nil
  })
  return res, dberr
}

func (db *DB) insertMessage(tx *sql.Tx, boxID int, msg *Message) error {
  // Insert an empty row in order to get/reserve the next message ID.
  res, err := tx.Exec(`
    insert into message(
      id,
      mailbox_id,
      size,
      created,
      path
    ) values (?, ?, ?, ?, ?)`,
    msg.ID, boxID, msg.Size, msg.Created, msg.Path)
  if err != nil {
    return fmt.Errorf("inserting mail into database: %v", err)
  }

  rowID, err := res.LastInsertId()
  if err != nil {
    return fmt.Errorf("getting inserted mail ID: %v", err)
  }
  // TODO need to think carefully about all the int conversions going on
  msg.RowID = int(rowID)

  err = db.addHeaders(tx, msg.RowID, msg.Headers)
  if err != nil {
    return err
  }

  err = db.addFlags(tx, msg.RowID, msg.Flags)
  if err != nil {
    return err
  }
  return nil
}

func (db *DB) nextID(tx *sql.Tx, mailbox string) (boxID, msgID int, err error) {
  // TODO this is going outside the transaction. how exactly does transaction
  //      timing work in sqlite? does this block all other transactions from starting?
  box, err := db.MailboxByName(mailbox)
  if err != nil {
    return 0, 0, err
  }
  return box.ID, box.NextMessageID, nil
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

func saveMessageBody(body io.Reader, out io.Writer) (h Headers, size int, err error) {
  // Limit the size of the message body.
  mr := &maxReader{R: body, N: MaxBodyBytes}
  sr := &sizeReader{R: mr}
  // tee because we need to parse the headers while copying to the output.
  r := io.TeeReader(sr, out)

  // Parse the message headers.
  m, err := mail.ReadMessage(r)
  if err != nil {
    return nil, 0, fmt.Errorf("parsing message headers: %v", err)
  }

  // Copy the data to the file.
  //
  // The ioutil.Discard looks weird here, but the TeeReader above copies
  // data to the output.
  _, err = io.Copy(ioutil.Discard, r)
  if err != nil {
    if err == errByteLimitReached {
      return nil, 0, fmt.Errorf("message body is too big. max is %d bytes.", MaxBodyBytes)
    }
    return nil, 0, fmt.Errorf("writing message body file: %v", err)
  }
  return Headers(m.Header), sr.N, nil
}
