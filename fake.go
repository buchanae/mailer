package main

type fake struct {}

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
  return nil
}

func (*fake) StartTLS() error {
  return nil
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
      "UIDVALIDITY": 4,
      "UNSEEN": 0,
    },
  }, nil
}

func (*fake) Fetch(*FetchRequest) (*FetchResponse, error) {
  return &FetchResponse{}, nil
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
  return &FetchResponse{}, nil
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
