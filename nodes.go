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
