package model

import (
  "io"
  "bytes"
  "fmt"
  "github.com/buchanae/mailer/imap"
)

func (db *DB) Search(cmd *imap.SearchCommand) ([]int, error) {

  buf := &bytes.Buffer{}
  b := &builder{
    Writer: buf,
  }
  b.expr("select distinct(msg.id) from message as msg")
  b.expr("join header on msg.row_id = header.message_row_id where")

  // TODO implement charset handling
  err := buildSearchQuery(b, &imap.GroupKey{Keys: cmd.Keys})
  if err != nil {
    return nil, err
  }

  q := buf.String()
  fmt.Println(q, b.args)
  rows, err := db.db.Query(q, b.args...)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var ids []int
  for rows.Next() {
    var id int
    err := rows.Scan(&id)
    if err != nil {
      return nil, err
    }
    ids = append(ids, id)
  }

  if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("loading message range: %v", err)
  }
  return ids, nil
}

type builder struct {
  io.Writer
  args []interface{}
}
func (b *builder) expr(s string, args ...interface{}) {
  fmt.Fprint(b, " ")
  fmt.Fprint(b, s)
  fmt.Fprint(b, " ")
  b.args = append(b.args, args...)
}

func buildSearchQuery(b *builder, key imap.SearchKey) error {
  switch z := key.(type) {

  case *imap.GroupKey:
    b.expr("(")
    started := false
    for _, k := range z.Keys {
      if started {
        b.expr("and")
      }
      err := buildSearchQuery(b, k)
      if err != nil {
        return err
      }
      started = true
    }
    b.expr(")")
    return nil

  case *imap.OrKey:
    err := buildSearchQuery(b, z.Arg1)
    if err != nil {
      return err
    }
    b.expr("or")
    err = buildSearchQuery(b, z.Arg2)
    if err != nil {
      return err
    }
    return nil

  case *imap.NotKey:
    b.expr("not (")
    err := buildSearchQuery(b, z.Arg)
    if err != nil {
      return err
    }
    b.expr(")")
    return nil

  case *imap.StatusKey:
    switch z.Name {
    // TODO
    case "all":
      // Apparently "all" means all messages in the mailbox, so there's
      // nothing to query for here. Seems silly?
    case  "answered":
      b.expr("answered = 1")
    case "unanswered":
      b.expr("answered = 0")
    case  "deleted":
      b.expr("deleted = 1")
    case  "undeleted":
      b.expr("deleted = 0")
    case  "flagged":
      b.expr("flagged = 1")
    case  "unflagged":
      b.expr("flagged = 0")
    case  "recent":
      b.expr("recent = 1")
    case  "new":
      b.expr("recent = 1 and seen = 0")
    case  "old":
      b.expr("recent = 0")
    case  "seen":
      b.expr("seen = 1")
    case  "unseen":
      b.expr("seen = 0")
    case  "draft":
      b.expr("draft = 1")
    case  "undraft":
      b.expr("draft = 0")
    default:
      return fmt.Errorf("unknown status key %q", z.Name)
    }

  case *imap.FieldKey:
    switch z.Name {
    case "bcc", "cc", "from", "subject", "to":
      arg := "%" + z.Arg + "%"
      b.expr("header.key = ? and header.value like ?", z.Name, arg)
    // TODO
    //case  "body":
    //case  "text":
    //case "keyword":
    //case  "unkeyword":
    default:
      return fmt.Errorf("unknown field key %q", z.Name)
    }

  case *imap.HeaderKey:
    arg := "%" + z.Arg + "%"
    b.expr("header.key = ? and header.value like ?", z.Name, arg)

  case *imap.DateKey:
    switch z.Name {
    case "before":
      b.expr("msg.created < ?", z.Arg)
    case  "since":
      b.expr("msg.created > ?", z.Arg)
    // TODO
    //case  "on":
    //case "sentbefore":
    //case  "senton":
    //case  "sentsince":
    default:
      return fmt.Errorf("unknown date key %q", z.Name)
    }

  case *imap.SizeKey:
    switch z.Name {
    case "larger":
      b.expr("msg.size > ?", z.Arg)
    case  "smaller":
      b.expr("msg.size < ?", z.Arg)
    default:
      return fmt.Errorf("unknown size key %q", z.Name)
    }

  case *imap.UIDKey:
    for _, seq := range z.Seqs {
      if seq.IsRange && seq.End > seq.Start {
        b.expr("msg.id >= ? and msg.id <= ?", seq.Start, seq.End)
      } else {
        b.expr("msg.id >= ?", seq.Start)
      }
    }

  case *imap.SequenceKey:
    return fmt.Errorf("sequence key is not supported")
  default:
    return fmt.Errorf("unknown search key")
  }
  return nil
}
