package stgxui

var (
	Commands = make(map[string]*Command, 0)
)

type Command struct {
	Display string
	Name    string
	command string
	call    func(*Window)
}

func (cmd *Command) Exec(stw *Window) {
	cmd.call(stw)
}
