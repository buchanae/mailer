package imap

type Command interface {
  IMAPTag() string
}

type UnknownCommand struct { Tag string }
type CapabilityCommand struct { Tag string }
type LogoutCommand struct { Tag string }
type NoopCommand struct { Tag string }
type StartTLSCommand struct { Tag string }
type CheckCommand struct { Tag string }
type CloseCommand struct { Tag string }
type ExpungeCommand struct { Tag string }

type UIDFetchCommand struct {
  *FetchCommand
}
type UIDStoreCommand struct {
  *StoreCommand
}
type UIDSearchCommand struct {
  *SearchCommand
}
type UIDCopyCommand struct {
  *CopyCommand
}

type LoginCommand struct {
  Tag string
  Username, Password string
}

type AuthenticateCommand struct {
  Tag string
  AuthType string
}

type CreateCommand struct {
  Tag string
  Mailbox string
}

type RenameCommand struct {
  Tag string
  From, To string
}

type DeleteCommand struct {
  Tag string
  Mailbox string
}

type ListCommand struct {
  Tag string
  Mailbox string
  Query string
}

type LsubCommand struct {
  Tag string
  Mailbox string
  Query string
}

type SubscribeCommand struct {
  Tag string
  Mailbox string
}

type UnsubscribeCommand struct {
  Tag string
  Mailbox string
}

type SelectCommand struct {
  Tag string
  Mailbox string
  Flags Flags
}

type ExamineCommand struct {
  Tag string
  Mailbox string
}

type StatusCommand struct {
  Tag string
  Mailbox string
  Attrs []string
}

type FetchCommand struct {
  Tag string
  Seqs []Sequence
  Attrs []*fetchAttrNode
}

type SearchCommand struct {
  Tag string
  Charset string
  Keys []searchKeyNode
}

type CopyCommand struct {
  Tag string
  Mailbox string
  Seqs []Sequence
}

type StoreCommand struct {
  Tag string

  plusMinus string
  seqs []Sequence
  key string
  flags []string
}

type AppendCommand struct {
  Tag string
}

func (x *UnknownCommand) IMAPTag() string { return x.Tag }
func (x *CapabilityCommand) IMAPTag() string { return x.Tag }
func (x *LogoutCommand) IMAPTag() string { return x.Tag }
func (x *NoopCommand) IMAPTag() string { return x.Tag }
func (x *StartTLSCommand) IMAPTag() string { return x.Tag }
func (x *CheckCommand) IMAPTag() string { return x.Tag }
func (x *CloseCommand) IMAPTag() string { return x.Tag }
func (x *ExpungeCommand) IMAPTag() string { return x.Tag }
func (x *LoginCommand) IMAPTag() string { return x.Tag }
func (x *CreateCommand) IMAPTag() string { return x.Tag }
func (x *DeleteCommand) IMAPTag() string { return x.Tag }
func (x *ExamineCommand) IMAPTag() string { return x.Tag }
func (x *ListCommand) IMAPTag() string { return x.Tag }
func (x *LsubCommand) IMAPTag() string { return x.Tag }
func (x *RenameCommand) IMAPTag() string { return x.Tag }
func (x *SelectCommand) IMAPTag() string { return x.Tag }
func (x *SubscribeCommand) IMAPTag() string { return x.Tag }
func (x *UnsubscribeCommand) IMAPTag() string { return x.Tag }
func (x *StatusCommand) IMAPTag() string { return x.Tag }
func (x *AuthenticateCommand) IMAPTag() string { return x.Tag }
func (x *FetchCommand) IMAPTag() string { return x.Tag }
func (x *CopyCommand) IMAPTag() string { return x.Tag }
func (x *StoreCommand) IMAPTag() string { return x.Tag }
func (x *SearchCommand) IMAPTag() string { return x.Tag }
func (x *AppendCommand) IMAPTag() string { return x.Tag }
