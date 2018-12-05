package model

import (
  "database/sql"
  "fmt"
)

func (db *DB) CreateMailbox(name string) error {
  _, err := db.db.Exec("insert into mailbox(name, next_message_id) values(?, 1)", name)
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

func (db *DB) ListMailboxes() ([]*Mailbox, error) {

  var boxes []*Mailbox

  rows, err := db.db.Query("select id, name, next_message_id from mailbox")
  if err != nil {
    return nil, fmt.Errorf("loading mailboxes from database: %v", err)
  }
  defer rows.Close()

  for rows.Next() {
    box := &Mailbox{}
    err := rows.Scan(&box.ID, &box.Name, &box.NextMessageID)
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
  q := "select id, next_message_id from mailbox where name = ?"
  row := db.db.QueryRow(q, name)
  err := row.Scan(&box.ID, &box.NextMessageID)
  if err == sql.ErrNoRows {
    return nil, fmt.Errorf("no mailbox named %q", name)
  }
  if err != nil {
    return nil, fmt.Errorf("finding mailbox by name: %v", err)
  }
  return box, nil
}

func (db *DB) MessageCount(mailbox string) (int, error) {
  var count int

  row := db.db.QueryRow(
    `select count(message.id)
    from message
    join mailbox
    on message.mailbox_id = mailbox.id
    where mailbox.name = ?`,
    mailbox)

  err := row.Scan(&count)
  if err != nil {
    return 0, fmt.Errorf("database error: getting message count: %v", err)
  }
  return count, nil
}

func (db *DB) RecentCount(mailbox string) (int, error) {
  var count int

  row := db.db.QueryRow(
    `select count(message.id)
    from message
    join mailbox
    on message.mailbox_id = mailbox.id
    where mailbox.name = ?
    and message.recent = 1`,
    mailbox)

  err := row.Scan(&count)
  if err != nil {
    return 0, fmt.Errorf("database error: getting recent message count: %v", err)
  }
  return count, nil
}

func (db *DB) UnseenCount(mailbox string) (int, error) {
  var count int

  row := db.db.QueryRow(
    `select count(message.id)
    from message
    join mailbox
    on message.mailbox_id = mailbox.id
    where mailbox.name = ?
    and message.seen = 0`,
    mailbox)

  err := row.Scan(&count)
  if err != nil {
    return 0, fmt.Errorf("database error: getting unseen message count: %v", err)
  }
  return count, nil
}
