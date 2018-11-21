package main

type node interface {
  isNode()
}

type strNode struct {
  str string
}

type numNode struct {
  num int
}

type literalNode struct {
  size int
  content string
}

type tagNode struct {
  tag string
}

type simpleCmd struct {
  tag, name string
}

type loginCmd struct {
  tag, user, pass string
}

type createCmd struct {
  tag, mailbox string
}

type deleteCmd struct {
  tag, mailbox string
}

type examineCmd struct {
  tag, mailbox string
}

type listCmd struct {
  tag, mailbox, query string
}

type lsubCmd struct {
  tag, mailbox, query string
}

type renameCmd struct {
  tag, from, to string
}

type selectCmd struct {
  tag, mailbox string
}

type subscribeCmd struct {
  tag, mailbox string
}

type unsubscribeCmd struct {
  tag, mailbox string
}

type statusCmd struct {
  tag, mailbox string
  attrs []string
}

type authCmd struct {
  tag, authType string
}

type fetchCmd struct {
  tag string
  seqs []seq
  attrs []*fetchAttrNode
}

type sectionNode struct {
  msg string
  headerList []string
}

type copyCmd struct {
  tag, mailbox string
  seqs []seq
}

type storeCmd struct {
  tag string
  plusMinus string
  seqs []seq
  key string
  flags []string
}

type searchCmd struct {
  charset string
  keys []searchKeyNode
}

type nstringNode struct {
  isNil bool
  str string
}

type addressNode struct {
  name, adl, mailbox, host *nstringNode
}

func (*strNode) isNode() {}
func (*numNode) isNode() {}
func (*literalNode) isNode() {}
func (*simpleCmd) isNode() {}
func (*loginCmd) isNode() {}
func (*createCmd) isNode() {}
func (*deleteCmd) isNode() {}
func (*examineCmd) isNode() {}
func (*listCmd) isNode() {}
func (*lsubCmd) isNode() {}
func (*renameCmd) isNode() {}
func (*selectCmd) isNode() {}
func (*subscribeCmd) isNode() {}
func (*unsubscribeCmd) isNode() {}
func (*statusCmd) isNode() {}
func (*authCmd) isNode() {}
func (*fetchCmd) isNode() {}
func (*copyCmd) isNode() {}
func (*storeCmd) isNode() {}

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
