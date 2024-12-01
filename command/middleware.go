package command

import (
	"context"
	"fmt"
)

func MiddlewareLoggedIn(state *State, cmd Command) error {
	user, err := state.DB.GetUser(context.Background(), state.Cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("couldn't get user: %w", err)
	}

	state.User = &user

	return nil
}
