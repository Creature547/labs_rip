package currentuser

import "github.com/gin-gonic/gin"

// ID belongs to this HTTP request. Zero denotes a guest.
func ID(c *gin.Context) uint {
	value, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	id, _ := value.(uint)
	return id
}
