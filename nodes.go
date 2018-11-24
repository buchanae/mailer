package main

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

type unknownRequest struct {
  tag string
}

type uidFetch struct {
  *FetchRequest
}
type uidStore struct {
  *StoreRequest
}
type uidSearch struct {
  *SearchRequest
}
type uidCopy struct {
  *CopyRequest
}

type capabilityRequest struct { tag string }
type logoutRequest struct { tag string }
type noopRequest struct { tag string }
type startTLSRequest struct { tag string }
type checkRequest struct { tag string }
type closeRequest struct { tag string }
type expungeRequest struct { tag string }

type commandI interface {
  requestTag() string
}
func (x *unknownRequest) requestTag() string { return x.tag }
func (x *capabilityRequest) requestTag() string { return x.tag }
func (x *logoutRequest) requestTag() string { return x.tag }
func (x *noopRequest) requestTag() string { return x.tag }
func (x *startTLSRequest) requestTag() string { return x.tag }
func (x *checkRequest) requestTag() string { return x.tag }
func (x *closeRequest) requestTag() string { return x.tag }
func (x *expungeRequest) requestTag() string { return x.tag }
func (x *LoginRequest) requestTag() string { return x.tag }
func (x *CreateRequest) requestTag() string { return x.tag }
func (x *DeleteRequest) requestTag() string { return x.tag }
func (x *ExamineRequest) requestTag() string { return x.tag }
func (x *ListRequest) requestTag() string { return x.tag }
func (x *LsubRequest) requestTag() string { return x.tag }
func (x *RenameRequest) requestTag() string { return x.tag }
func (x *SelectRequest) requestTag() string { return x.tag }
func (x *SubscribeRequest) requestTag() string { return x.tag }
func (x *UnsubscribeRequest) requestTag() string { return x.tag }
func (x *StatusRequest) requestTag() string { return x.tag }
func (x *AuthenticateRequest) requestTag() string { return x.tag }
func (x *FetchRequest) requestTag() string { return x.tag }
func (x *CopyRequest) requestTag() string { return x.tag }
func (x *StoreRequest) requestTag() string { return x.tag }
func (x *SearchRequest) requestTag() string { return x.tag }
func (x *AppendRequest) requestTag() string { return x.tag }

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
