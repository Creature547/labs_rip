package currentuser

import "testing"

func TestGetIsSingleton(t *testing.T) {
	first := Get()
	second := Get()
	if first != second {
		t.Fatal("Get must return the same singleton instance")
	}
	if first.ID != FixedUserID {
		t.Fatalf("unexpected fixed user id: %d", first.ID)
	}
}
