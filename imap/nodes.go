package imap

type Flags struct {
  Seen, Answered, Flagged, Deleted, Draft, Recent bool
}

type literalNode struct {
  size int
  content string
}

type fetchAttrNode struct {
	name string
  peek bool
	sec  *sectionNode
}

type sectionNode struct {
  msg string
  headerList []string
}

type nstringNode struct {
  isNil bool
  str string
}

type addressNode struct {
  name, adl, mailbox, host *nstringNode
}

type searchKeyNode interface {
  isSearchKey()
}

type bccSearchKey struct {
  arg string
}

type bodySearchKey struct {
  arg string
}

type ccSearchKey struct {
  arg string
}

type fromSearchKey struct {
  arg string
}

type toSearchKey struct {
  arg string
}

type simpleSearchKey struct {
  name string
}

type subjectSearchKey struct {
  name string
}

type notSearchKey struct {
  arg searchKeyNode
}

type orSearchKey struct {
  arg1, arg2 searchKeyNode
}

type textSearchKey struct {
  arg string
}

func (*bccSearchKey) isSearchKey() {}
func (*bodySearchKey) isSearchKey() {}
func (*ccSearchKey) isSearchKey() {}
func (*fromSearchKey) isSearchKey() {}
func (*toSearchKey) isSearchKey() {}
func (*subjectSearchKey) isSearchKey() {}
func (*simpleSearchKey) isSearchKey() {}
func (*textSearchKey) isSearchKey() {}
func (*notSearchKey) isSearchKey() {}
func (*orSearchKey) isSearchKey() {}
