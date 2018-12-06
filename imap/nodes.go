package imap

import (
  "time"
)

type Partial struct {
  Offset int
  Limit int
}

type FetchAttr struct {
	Name string
  Headers []string
  Part []int
  Partial *Partial
}

type SearchKey interface {
  isSearchKey()
}

type GroupKey struct {
  Keys []SearchKey
}

type StatusKey struct {
  Name string
}

type FieldKey struct {
  Name string
  Arg string
}

type OrKey struct {
  Arg1, Arg2 SearchKey
}

type NotKey struct {
  Arg SearchKey
}

type HeaderKey struct {
  Name string
  Arg string
}

type DateKey struct {
  Name string
  Arg time.Time
}

type SizeKey struct {
  Name string
  Arg int
}

type UIDKey struct {
  Seqs []Sequence
}

type SequenceKey struct {
  Seqs []Sequence
}

func (*FieldKey) isSearchKey() {}
func (*StatusKey) isSearchKey() {}
func (*OrKey) isSearchKey() {}
func (*NotKey) isSearchKey() {}
func (*HeaderKey) isSearchKey() {}
func (*DateKey) isSearchKey() {}
func (*SizeKey) isSearchKey() {}
func (*UIDKey) isSearchKey() {}
func (*SequenceKey) isSearchKey() {}
func (*GroupKey) isSearchKey() {}
