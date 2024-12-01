package command

import (
	"errors"
	"log"

	"github.com/zokkis/gator/internal/config"
	"github.com/zokkis/gator/internal/database"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	RegisteredCommands map[string][]func(*State, Command) error
}

type State struct {
	DB   *database.Queries
	Cfg  *config.Config
	User *database.User
}

func (commands *Commands) Register(name string, funcs ...(func(*State, Command) error)) {
	commands.RegisteredCommands[name] = funcs
}

func (commands *Commands) Run(state *State, cmd Command) error {
	funcs, ok := commands.RegisteredCommands[cmd.Name]
	if !ok {
		return errors.New("command not found")
	}
	for _, function := range funcs {
		if err := function(state, cmd); err != nil {
			return err
		}
	}
	return nil
}

func (commands *Commands) MustRun(state *State, cmd Command) {
	err := commands.Run(state, cmd)
	if err != nil {
		log.Fatalln(err)
	}
}
