package data

import (
	"fmt"

	"vim-guys.theprimeagen.tv/auth-proxy/pkg/config"
)

// UserMapping represents the structure of the user_mapping table
type UserMapping struct {
    UserID     string `db:"userId"`
    UUID string `db:"uuid"`
}

func (u *UserMapping) String() string {
	return fmt.Sprintf("UserId: %s -- UUID: %s", u.UserID, u.UUID)
}

func AccountExists(ctx config.ProxyContext, uuid string) bool {
	query := "SELECT userId, uuid FROM user_mapping WHERE uuid = ?"
	var mapping UserMapping
	err := ctx.DB.Get(&mapping, query, uuid)
	if err != nil {
		ctx.Logger.Debug("unable to get user token", "error", err)
	}

	return err == nil
}
