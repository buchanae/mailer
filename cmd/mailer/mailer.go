package main

import (
  "github.com/spf13/cobra"
)

var root = &cobra.Command{
  Use: "mailer",
}

func init() {
  root.AddCommand(createMessage)
  root.AddCommand(getMessage)

  root.AddCommand(createMailbox)
  root.AddCommand(renameMailbox)
  root.AddCommand(deleteMailbox)
  root.AddCommand(listBoxes)
}

func main() {
  root.Execute()
}
