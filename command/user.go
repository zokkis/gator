package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zokkis/gator/internal/database"
)

func Login(state *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, err := state.DB.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("couldn't get user: %w", err)
	}

	err = state.Cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User switched successfully!")
	return nil
}

func Register(state *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	currentTime := time.Now()
	user, err := state.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return fmt.Errorf("couldn't register current user: %w", err)
	}

	err = state.Cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("User registered successfully: %v\n", user)
	return nil
}

func ListUsers(state *State, cmd Command) error {
	users, err := state.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't get users: %w", err)
	}

	for _, user := range users {
		current := ""
		if user.Name == state.Cfg.CurrentUserName {
			current = " (current)"
		}
		fmt.Printf("* %s%s\n", user.Name, current)
	}

	return nil
}

func Reset(state *State, cmd Command) error {
	err := state.DB.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't delete users: %w", err)
	}

	fmt.Println("Database reset successfully!")
	return nil
}
