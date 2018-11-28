package main

import (
  "github.com/spf13/cobra"
)

var root = &cobra.Command{
  Use: "mailer",
}

func init() {
  root.AddCommand(createMailbox)
  root.AddCommand(createMessage)
  root.AddCommand(getMessage)
  root.AddCommand(listBoxes)
}

func main() {
  root.Execute()
}
