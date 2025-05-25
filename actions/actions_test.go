package actions

import (
	"os"
	"sync"
	"testing"

	"github.com/gobuffalo/suite/v4"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	// Ensure we're running in test environment to disable CSRF
	os.Setenv("GO_ENV", "test")

	// Reset the app instance so it gets recreated with test environment
	appOnce = sync.Once{}
	app = nil

	action, err := suite.NewActionWithFixtures(App(), os.DirFS("../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}
