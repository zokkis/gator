package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // postgres driver

	"github.com/zokkis/gator/command"
	"github.com/zokkis/gator/internal/config"
	"github.com/zokkis/gator/internal/database"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	programState := &command.State{
		Cfg: &cfg,
		DB:  database.New(db),
	}

	cmds := command.Commands{
		RegisteredCommands: make(map[string][]func(*command.State, command.Command) error),
	}
	cmds.Register("login", command.Login)
	cmds.Register("register", command.Register)
	cmds.Register("users", command.ListUsers)
	cmds.Register("reset", command.Reset)
	cmds.Register("agg", command.FetchFeed)
	cmds.Register("addfeed", command.MiddlewareLoggedIn, command.AddFeed)
	cmds.Register("feeds", command.ListFeeds)
	cmds.Register("follow", command.MiddlewareLoggedIn, command.FollowFeed)
	cmds.Register("following", command.MiddlewareLoggedIn, command.ListFollowing)

	if len(os.Args) < 2 {
		fmt.Println("Usage: cli <command> [args...]")
		return
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	cmds.MustRun(programState, command.Command{Name: cmdName, Args: cmdArgs})
}
