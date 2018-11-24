package main

import (
  "log"
  "net/mail"
  "bytes"
  "github.com/kr/pretty"
)

func main() {
  body := "From: \"test from\" <from@nobody.com>\r\nTo: <buchanae@gmail.com>\r\nSubject: Help with your email server\r\nDate: Wed, 03 Oct 2018 21:08:41 -0600\r\nMessage-ID: <a438f673-6ec7-47b1-b291-488d11ed8d10@las1s04mta912.xt.local>\r\n\r\n"
  r := bytes.NewBufferString(body)
  msg, err := mail.ReadMessage(r)
  if err != nil {
    log.Println(err)
    return
  }
  pretty.Println(msg.Header)
}
