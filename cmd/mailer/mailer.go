package main

import (
  "github.com/spf13/cobra"
  "github.com/buchanae/mailer"
)

var root = &cobra.Command{
  Use: "mailer",
}

var run = &cobra.Command{
  Use: "run",
  Args: cobra.ExactArgs(0),
  Run: func(cmd *cobra.Command, args []string) {
    mailer.Run()
  },
}

func init() {
  root.AddCommand(createMessage)
  root.AddCommand(getMessage)
  root.AddCommand(run)

  root.AddCommand(createMailbox)
  root.AddCommand(renameMailbox)
  root.AddCommand(deleteMailbox)
  root.AddCommand(listBoxes)
}

func main() {
  root.Execute()
}
