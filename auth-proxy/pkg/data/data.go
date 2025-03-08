package data

import "fmt"

// UserMapping represents the structure of the user_mapping table
type UserMapping struct {
    UserID     string `db:"userId"`
    UUID string `db:"uuid"`
}

func (u *UserMapping) String() string {
	return fmt.Sprintf("UserId: %s -- UUID: %s", u.UserID, u.UUID)
}


