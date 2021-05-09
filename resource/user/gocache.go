package user

import (
	"fmt"

	"github.com/lolmourne/go-websocket/model"
	"github.com/patrickmn/go-cache"
)

func (acr *AuthCliResource) GetUserByID(userID int64) *model.User {
	usr, ok := acr.goc.Get(fmt.Sprintf("usr:%d", userID))

	if !ok {
		userCli := acr.authClient.GetUserByID(userID)
		if userCli == nil {
			return nil
		}
		user := &model.User{
			UserID:     userCli.UserID,
			ProfilePic: userCli.ProfilePic,
			Username:   userCli.Username,
			CreatedAt:  userCli.CreatedAt,
		}
		acr.goc.Set(fmt.Sprintf("usr:%d", userID), user, cache.DefaultExpiration)

		return user
	}

	return usr.(*model.User)
}
