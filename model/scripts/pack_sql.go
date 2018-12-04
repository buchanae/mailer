package main

import (
  "fmt"
  "path/filepath"
  "os"
  "log"
  "io"
  "io/ioutil"
)

func main() {
  files, err := ioutil.ReadDir("./sql")
  if err != nil {
    log.Fatal(err)
  }

  fh, err := os.Create("packed.go")
  if err != nil {
    log.Fatal(err)
  }
  defer fh.Close()
  fmt.Fprintln(fh, `package model

// DO NOT EDIT: generated by ./scripts/pack_sql.go
`)

  fmt.Fprintln(fh, "var packed = `")
  for _, f := range files {
    path := filepath.Join("sql", f.Name())
    log.Println("packing:", path)

    f, err := os.Open(path)
    if err != nil {
      log.Fatal(err)
    }
    defer f.Close()

    _, err = io.Copy(fh, f)
    if err != nil {
      log.Fatal(err)
    }

    fmt.Fprintln(fh, "")
  }
  fmt.Fprintln(fh, "`")
}
