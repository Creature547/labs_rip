package currentuser

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestUsersAreRequestScoped(t *testing.T) {
	a, _ := gin.CreateTestContext(httptest.NewRecorder())
	b, _ := gin.CreateTestContext(httptest.NewRecorder())
	a.Set("user_id", uint(42))
	if ID(a) != 42 || ID(b) != 0 {
		t.Fatal("user leaked between requests")
	}
}
