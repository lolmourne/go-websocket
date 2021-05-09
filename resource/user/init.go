package user

import (
	"time"

	authClient "github.com/lolmourne/go-accounts/client/userauth"
	"github.com/lolmourne/go-websocket/model"
	"github.com/patrickmn/go-cache"
)

type IResource interface {
	GetUserByID(userID int64) *model.User
}

type AuthCliResource struct {
	authClient authClient.ClientItf
	goc        *cache.Cache
}

func NewAuthCliRsc(authClient authClient.ClientItf, ttl, purgeTime time.Duration) IResource {
	c := cache.New(purgeTime*time.Minute, purgeTime*time.Minute)
	return &AuthCliResource{
		goc:        c,
		authClient: authClient,
	}
}
