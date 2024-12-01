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
	RegisteredCommands map[string]func(*State, Command) error
}

type State struct {
	DB  *database.Queries
	Cfg *config.Config
}

func (commands *Commands) Register(name string, function func(*State, Command) error) {
	commands.RegisteredCommands[name] = function
}

func (commands *Commands) Run(state *State, cmd Command) error {
	function, ok := commands.RegisteredCommands[cmd.Name]
	if !ok {
		return errors.New("command not found")
	}
	return function(state, cmd)
}

func (commands *Commands) MustRun(state *State, cmd Command) {
	err := commands.Run(state, cmd)
	if err != nil {
		log.Fatalln(err)
	}
}
