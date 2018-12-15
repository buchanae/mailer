package main

import cli "github.com/buchanae/cli"
import mailer "github.com/buchanae/mailer"

func specs() []cli.Spec {
	return []cli.Spec{
		&createMailboxSpec{
			opt: DefaultOpt(),
		},
		&deleteMailboxSpec{
			opt: DefaultOpt(),
		},
		&renameMailboxSpec{
			opt: DefaultOpt(),
		},
		&getMessageSpec{
			opt: DefaultOpt(),
		},
		&getMailboxSpec{
			opt: DefaultOpt(),
		},
		&createMessageSpec{
			opt: DefaultOpt(),
		},
		&listMailboxesSpec{
			opt: DefaultOpt(),
		},
		&runSpec{
			opt: mailer.DefaultServerOpt(),
		},
		&runDevSpec{
			opt: DefaultDevServerOpt(),
		},
	}
}

type createMailboxSpec struct {
	cmd  *cli.Cmd
	opt  Opt
	args struct {
		arg0 string
	}
}

func (cmd *createMailboxSpec) Run() {
	CreateMailbox(
		cmd.opt,
		cmd.args.arg0,
	)
}

func (cmd *createMailboxSpec) Cmd() *cli.Cmd {
	if cmd.cmd != nil {
		return cmd.cmd
	}
	cmd.cmd = &cli.Cmd{
		RawName: "CreateMailbox",
		RawDoc:  "",
		Args: []*cli.Arg{
			{
				Name:     "name",
				Type:     "string",
				Variadic: false,
				Value:    &cmd.args.arg0,
			},
		},
		Opts: []*cli.Opt{
			{
				Key:          []string{"DB", "Path"},
				RawDoc:       "",
				Value:        &cmd.opt.DB.Path,
				DefaultValue: cmd.opt.DB.Path,
				Type:         "string",
				Short:        "",
			},
		},
	}
	cli.Enrich(cmd.cmd)
	return cmd.cmd
}

type deleteMailboxSpec struct {
	cmd  *cli.Cmd
	opt  Opt
	args struct {
		arg0 string
	}
}

func (cmd *deleteMailboxSpec) Run() {
	DeleteMailbox(
		cmd.opt,
		cmd.args.arg0,
	)
}

func (cmd *deleteMailboxSpec) Cmd() *cli.Cmd {
	if cmd.cmd != nil {
		return cmd.cmd
	}
	cmd.cmd = &cli.Cmd{
		RawName: "DeleteMailbox",
		RawDoc:  "",
		Args: []*cli.Arg{
			{
				Name:     "name",
				Type:     "string",
				Variadic: false,
				Value:    &cmd.args.arg0,
			},
		},
		Opts: []*cli.Opt{
			{
				Key:          []string{"DB", "Path"},
				RawDoc:       "",
				Value:        &cmd.opt.DB.Path,
				DefaultValue: cmd.opt.DB.Path,
				Type:         "string",
				Short:        "",
			},
		},
	}
	cli.Enrich(cmd.cmd)
	return cmd.cmd
}

type renameMailboxSpec struct {
	cmd  *cli.Cmd
	opt  Opt
	args struct {
		arg0 string
		arg1 string
	}
}

func (cmd *renameMailboxSpec) Run() {
	RenameMailbox(
		cmd.opt,
		cmd.args.arg0,
		cmd.args.arg1,
	)
}

func (cmd *renameMailboxSpec) Cmd() *cli.Cmd {
	if cmd.cmd != nil {
		return cmd.cmd
	}
	cmd.cmd = &cli.Cmd{
		RawName: "RenameMailbox",
		RawDoc:  "",
		Args: []*cli.Arg{
			{
				Name:     "from",
				Type:     "string",
				Variadic: false,
				Value:    &cmd.args.arg0,
			}, {
				Name:     "to",
				Type:     "string",
				Variadic: false,
				Value:    &cmd.args.arg1,
			},
		},
		Opts: []*cli.Opt{
			{
				Key:          []string{"DB", "Path"},
				RawDoc:       "",
				Value:        &cmd.opt.DB.Path,
				DefaultValue: cmd.opt.DB.Path,
				Type:         "string",
				Short:        "",
			},
		},
	}
	cli.Enrich(cmd.cmd)
	return cmd.cmd
}

type getMessageSpec struct {
	cmd  *cli.Cmd
	opt  Opt
	args struct {
		arg0 int
	}
}

func (cmd *getMessageSpec) Run() {
	GetMessage(
		cmd.opt,
		cmd.args.arg0,
	)
}

func (cmd *getMessageSpec) Cmd() *cli.Cmd {
	if cmd.cmd != nil {
		return cmd.cmd
	}
	cmd.cmd = &cli.Cmd{
		RawName: "GetMessage",
		RawDoc:  "",
		Args: []*cli.Arg{
			{
				Name:     "id",
				Type:     "int",
				Variadic: false,
				Value:    &cmd.args.arg0,
			},
		},
		Opts: []*cli.Opt{
			{
				Key:          []string{"DB", "Path"},
				RawDoc:       "",
				Value:        &cmd.opt.DB.Path,
				DefaultValue: cmd.opt.DB.Path,
				Type:         "string",
				Short:        "",
			},
		},
	}
	cli.Enrich(cmd.cmd)
	return cmd.cmd
}

type getMailboxSpec struct {
	cmd  *cli.Cmd
	opt  Opt
	args struct {
		arg0 string
	}
}

func (cmd *getMailboxSpec) Run() {
	GetMailbox(
		cmd.opt,
		cmd.args.arg0,
	)
}

func (cmd *getMailboxSpec) Cmd() *cli.Cmd {
	if cmd.cmd != nil {
		return cmd.cmd
	}
	cmd.cmd = &cli.Cmd{
		RawName: "GetMailbox",
		RawDoc:  "",
		Args: []*cli.Arg{
			{
				Name:     "name",
				Type:     "string",
				Variadic: false,
				Value:    &cmd.args.arg0,
			},
		},
		Opts: []*cli.Opt{
			{
				Key:          []string{"DB", "Path"},
				RawDoc:       "",
				Value:        &cmd.opt.DB.Path,
				DefaultValue: cmd.opt.DB.Path,
				Type:         "string",
				Short:        "",
			},
		},
	}
	cli.Enrich(cmd.cmd)
	return cmd.cmd
}

type createMessageSpec struct {
	cmd  *cli.Cmd
	opt  Opt
	args struct {
		arg0 string
		arg1 string
	}
}

func (cmd *createMessageSpec) Run() {
	CreateMessage(
		cmd.opt,
		cmd.args.arg0,
		cmd.args.arg1,
	)
}

func (cmd *createMessageSpec) Cmd() *cli.Cmd {
	if cmd.cmd != nil {
		return cmd.cmd
	}
	cmd.cmd = &cli.Cmd{
		RawName: "CreateMessage",
		RawDoc:  "",
		Args: []*cli.Arg{
			{
				Name:     "mailbox",
				Type:     "string",
				Variadic: false,
				Value:    &cmd.args.arg0,
			}, {
				Name:     "path",
				Type:     "string",
				Variadic: false,
				Value:    &cmd.args.arg1,
			},
		},
		Opts: []*cli.Opt{
			{
				Key:          []string{"DB", "Path"},
				RawDoc:       "",
				Value:        &cmd.opt.DB.Path,
				DefaultValue: cmd.opt.DB.Path,
				Type:         "string",
				Short:        "",
			},
		},
	}
	cli.Enrich(cmd.cmd)
	return cmd.cmd
}

type listMailboxesSpec struct {
	cmd  *cli.Cmd
	opt  Opt
	args struct {
	}
}

func (cmd *listMailboxesSpec) Run() {
	ListMailboxes(
		cmd.opt,
	)
}

func (cmd *listMailboxesSpec) Cmd() *cli.Cmd {
	if cmd.cmd != nil {
		return cmd.cmd
	}
	cmd.cmd = &cli.Cmd{
		RawName: "ListMailboxes",
		RawDoc:  "",
		Args:    []*cli.Arg{},
		Opts: []*cli.Opt{
			{
				Key:          []string{"DB", "Path"},
				RawDoc:       "",
				Value:        &cmd.opt.DB.Path,
				DefaultValue: cmd.opt.DB.Path,
				Type:         "string",
				Short:        "",
			},
		},
	}
	cli.Enrich(cmd.cmd)
	return cmd.cmd
}

type runSpec struct {
	cmd  *cli.Cmd
	opt  mailer.ServerOpt
	args struct {
	}
}

func (cmd *runSpec) Run() {
	Run(
		cmd.opt,
	)
}

func (cmd *runSpec) Cmd() *cli.Cmd {
	if cmd.cmd != nil {
		return cmd.cmd
	}
	cmd.cmd = &cli.Cmd{
		RawName: "Run",
		RawDoc:  "",
		Args:    []*cli.Arg{},
		Opts: []*cli.Opt{
			{
				Key:          []string{"SMTP", "Addr"},
				RawDoc:       "",
				Value:        &cmd.opt.SMTP.Addr,
				DefaultValue: cmd.opt.SMTP.Addr,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"SMTP", "Timeout"},
				RawDoc:       "",
				Value:        &cmd.opt.SMTP.Timeout,
				DefaultValue: cmd.opt.SMTP.Timeout,
				Type:         "time.Duration",
				Short:        "",
			}, {
				Key:          []string{"IMAP", "Addr"},
				RawDoc:       "",
				Value:        &cmd.opt.IMAP.Addr,
				DefaultValue: cmd.opt.IMAP.Addr,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"TLS", "Cert"},
				RawDoc:       "",
				Value:        &cmd.opt.TLS.Cert,
				DefaultValue: cmd.opt.TLS.Cert,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"TLS", "Key"},
				RawDoc:       "",
				Value:        &cmd.opt.TLS.Key,
				DefaultValue: cmd.opt.TLS.Key,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"DB", "Path"},
				RawDoc:       "",
				Value:        &cmd.opt.DB.Path,
				DefaultValue: cmd.opt.DB.Path,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"User", "Name"},
				RawDoc:       "",
				Value:        &cmd.opt.User.Name,
				DefaultValue: cmd.opt.User.Name,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"User", "Password"},
				RawDoc:       "",
				Value:        &cmd.opt.User.Password,
				DefaultValue: cmd.opt.User.Password,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"User", "NoAuth"},
				RawDoc:       "",
				Value:        &cmd.opt.User.NoAuth,
				DefaultValue: cmd.opt.User.NoAuth,
				Type:         "bool",
				Short:        "",
			}, {
				Key:          []string{"Debug", "ConnLog"},
				RawDoc:       "",
				Value:        &cmd.opt.Debug.ConnLog,
				DefaultValue: cmd.opt.Debug.ConnLog,
				Type:         "string",
				Short:        "",
			},
		},
	}
	cli.Enrich(cmd.cmd)
	return cmd.cmd
}

type runDevSpec struct {
	cmd  *cli.Cmd
	opt  DevServerOpt
	args struct {
	}
}

func (cmd *runDevSpec) Run() {
	RunDev(
		cmd.opt,
	)
}

func (cmd *runDevSpec) Cmd() *cli.Cmd {
	if cmd.cmd != nil {
		return cmd.cmd
	}
	cmd.cmd = &cli.Cmd{
		RawName: "RunDev",
		RawDoc:  "",
		Args:    []*cli.Arg{},
		Opts: []*cli.Opt{
			{
				Key:          []string{"SMTP", "Addr"},
				RawDoc:       "",
				Value:        &cmd.opt.SMTP.Addr,
				DefaultValue: cmd.opt.SMTP.Addr,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"SMTP", "Timeout"},
				RawDoc:       "",
				Value:        &cmd.opt.SMTP.Timeout,
				DefaultValue: cmd.opt.SMTP.Timeout,
				Type:         "time.Duration",
				Short:        "",
			}, {
				Key:          []string{"IMAP", "Addr"},
				RawDoc:       "",
				Value:        &cmd.opt.IMAP.Addr,
				DefaultValue: cmd.opt.IMAP.Addr,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"TLS", "Cert"},
				RawDoc:       "",
				Value:        &cmd.opt.TLS.Cert,
				DefaultValue: cmd.opt.TLS.Cert,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"TLS", "Key"},
				RawDoc:       "",
				Value:        &cmd.opt.TLS.Key,
				DefaultValue: cmd.opt.TLS.Key,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"DB", "Path"},
				RawDoc:       "",
				Value:        &cmd.opt.DB.Path,
				DefaultValue: cmd.opt.DB.Path,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"User", "Name"},
				RawDoc:       "",
				Value:        &cmd.opt.User.Name,
				DefaultValue: cmd.opt.User.Name,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"User", "Password"},
				RawDoc:       "",
				Value:        &cmd.opt.User.Password,
				DefaultValue: cmd.opt.User.Password,
				Type:         "string",
				Short:        "",
			}, {
				Key:          []string{"User", "NoAuth"},
				RawDoc:       "",
				Value:        &cmd.opt.User.NoAuth,
				DefaultValue: cmd.opt.User.NoAuth,
				Type:         "bool",
				Short:        "",
			}, {
				Key:          []string{"Debug", "ConnLog"},
				RawDoc:       "",
				Value:        &cmd.opt.Debug.ConnLog,
				DefaultValue: cmd.opt.Debug.ConnLog,
				Type:         "string",
				Short:        "",
			},
		},
	}
	cli.Enrich(cmd.cmd)
	return cmd.cmd
}

