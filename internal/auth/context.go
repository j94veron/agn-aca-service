package auth

import "github.com/gin-gonic/gin"

func FromGinContext(c *gin.Context) *User {
	u, ok := c.Get("user")
	if !ok {
		return nil
	}
	user, _ := u.(*User)
	return user
}
