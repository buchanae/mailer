package model

import (
  "fmt"
  "database/sql"
  "github.com/buchanae/mailer/imap"
)

func (db *DB) AddFlags(rowID int, flags []imap.Flag) error {
  return db.withTx(func(tx *sql.Tx) error {
    return db.addFlags(tx, rowID, flags)
  })
}

func (db *DB) RemoveFlags(rowID int, flags []imap.Flag) error {
  return db.withTx(func(tx *sql.Tx) error {
    return db.removeFlags(tx, rowID, flags)
  })
}

func (db *DB) ReplaceFlags(rowID int, remove, add []imap.Flag) error {
  return db.withTx(func(tx *sql.Tx) error {
    err := db.removeFlags(tx, rowID, remove)
    if err != nil {
      return err
    }

    err = db.addFlags(tx, rowID, add)
    if err != nil {
      return err
    }
    return nil
  })
}

func (db *DB) addFlags(tx *sql.Tx, rowID int, flags []imap.Flag) error {
  for _, flag := range flags {
    _, err := tx.Exec(
      "insert or ignore into flag (message_row_id, value) values (?, ?)",
      rowID, flag)
    if err != nil {
      return err
    }
  }
  return nil
}

func (db *DB) removeFlags(tx *sql.Tx, rowID int, flags []imap.Flag) error {
  for _, flag := range flags {
    _, err := tx.Exec(
      "delete from flag where message_row_id = ? and value = ?",
      rowID, flag)
    if err != nil {
      return err
    }
  }
  return nil
}

func (db *DB) loadFlags(msg *Message) error {
  rows, err := db.db.Query(`select value from flag where message_row_id = ?`, msg.RowID)
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
  rows, err := db.db.Query(`select key, value from header where message_row_id = ?`, msg.RowID)
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
