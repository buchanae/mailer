package main

import (
  "fmt"
  "github.com/buchanae/mailer/model"
)

type fake struct {
  Mailbox string
  db *model.DB
}

func (*fake) Noop() (*NoopResponse, error) {
  return &NoopResponse{}, nil
}

func (*fake) Check() error {
  return nil
}

func (*fake) Capability() (*CapabilityResponse, error) {
  return &CapabilityResponse{}, nil
}

func (*fake) Expunge() (*ExpungeResponse, error) {
  return &ExpungeResponse{}, nil
}

func (*fake) Login(*LoginRequest) error {
  return nil
}

func (*fake) Logout() error {
  return nil
}

func (*fake) Authenticate(*AuthenticateRequest) error {
    /*
    TODO auth is difficult because it's a multi-step
         challenge/response
    if z.authType == "PLAIN" {
      wr("+")
      tok := base64(r)
      crlf(r)
      log.Println("AUTH TOK", tok)
    }
    */
  return nil, fmt.Errorf("authenticate is not implemented")
}

func (*fake) StartTLS() error {
  return nil, fmt.Errorf("startTLS is not implemented")
}

func (*fake) Create(*CreateRequest) error {
  return nil
}

func (*fake) Rename(*RenameRequest) error {
  return nil
}

func (*fake) Delete(*DeleteRequest) error {
  return nil
}

func (*fake) List(r *ListRequest) (*ListResponse, error) {
  switch r.Query {
  case "":
    return &ListResponse{
      Items: []ListItem{
        {NoSelect: true, Delimiter: "/"},
      },
    }, nil

  case "*":
    return &ListResponse{
      Items: []ListItem{
        {Delimiter: "/", Name: "testone"},
      },
    }, nil
  }
  return &ListResponse{}, nil
}

func (*fake) Lsub(*LsubRequest) (*LsubResponse, error) {
  return &LsubResponse{}, nil
}

func (*fake) Subscribe(*SubscribeRequest) error {
  return nil
}

func (*fake) Unsubscribe(*UnsubscribeRequest) error {
  return nil
}

func (*fake) Select(*SelectRequest) (*SelectResponse, error) {
  return &SelectResponse{}, nil
}

func (*fake) Close() error {
  return nil
}

func (*fake) Examine(*ExamineRequest) (*ExamineResponse, error) {
  return &ExamineResponse{}, nil
}

func (*fake) Status(r *StatusRequest) (*StatusResponse, error) {
  return &StatusResponse{
    Mailbox: r.Mailbox,
    Counts: map[string]int{
      "MESSAGES": 1,
      "UIDNEXT": 6,
      "UIDVALIDITY": 1,
      "UNSEEN": 0,
    },
  }, nil
}

func (*fake) setSeen(msg *Message) {}

func (f *fake) Fetch(req *FetchRequest) (*FetchResponse, error) {
  resp := &FetchResponse{}

  for _, offset := range req.Seqs.Range() {
    msg, err := fake.db.MessageAtOffset(offset)
    if err != nil {
      return nil, fmt.Errorf("retrieving message range: %v", err)
    }

    item, err := fake.fetch(offset, msg, req)
    if err != nil {
      return nil, err
    }

    resp.Items = append(resp.Items, item)
  }

  return resp, nil
}

func (f *fake) fetch(id int, msg *Message, req *FetchRequest) (*FetchItem, error) {
  if shouldLoadText(req.Attrs) {
    text, err := fake.db.MessageText(msg.ID)
    if err != nil {
      return nil, err
    }
    msg.Text = text
  }

  if shouldSetSeen(req.Attrs) {
    err := f.setSeen(msg.ID)
    if err != nil {
      return nil, err
    }
  }
  return FetchItem{ID: id, Fields: fetchFields(msg, req.Attrs)}
}

func (*fake) Search(*SearchRequest) (*SearchResponse, error) {
  return &SearchResponse{}, nil
}

func (*fake) Copy(*CopyRequest) error {
  return nil
}

func (*fake) Store(*StoreRequest) (*StoreResponse, error) {
  return &StoreResponse{}, nil
}

func (*fake) Append(*AppendRequest) error {
  return nil
}

func (*fake) UIDFetch(*FetchRequest) (*FetchResponse, error) {
  resp := &FetchResponse{}

  for _, id := range req.Seqs.Range() {
    msg, err := fake.db.Message(id)
    if err != nil {
      return nil, fmt.Errorf("retrieving message range: %v", err)
    }

    item, err := fake.fetch(id, msg, req)
    if err != nil {
      return nil, err
    }

    resp.Items = append(resp.Items, item)
  }

  return resp, nil
}

func (*fake) UIDCopy(*CopyRequest) error {
  return nil
}

func (*fake) UIDStore(*StoreRequest) (*StoreResponse, error) {
  return &StoreResponse{}, nil
}

func (*fake) UIDSearch(*SearchRequest) (*SearchResponse, error) {
  return &SearchResponse{}, nil
}
