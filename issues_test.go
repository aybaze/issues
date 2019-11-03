package issues

import "testing"

func TestCheckContainsIfIssue(t *testing.T) {
	body := "This issue serves a collection of several issues related to HUD design\n\n- [ ] Movable windows\n- [ ] Resizable windows (#9)\n- [ ] Persistent window position across game-player (#7)\n- [ ]  Items should be link-able from inventory"

	newBody, status := CheckIfContainsIssue(body, "Something really awesome", 8)
	log.Printf("%s %d", newBody, status)
}
