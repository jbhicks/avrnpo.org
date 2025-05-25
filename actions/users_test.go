package actions

import (
	"fmt"
	"net/http"
	"time"

	"my_go_saas_template/models"
)

func (as *ActionSuite) Test_Users_New() {
	res := as.HTML("/users/new").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_Users_Create() {
	count, err := as.DB.Count("users")
	as.NoError(err)
	as.Equal(0, count)

	// Generate a unique email using timestamp to avoid any conflicts
	uniqueEmail := fmt.Sprintf("test-user-%d@example.com", time.Now().UnixNano())

	u := &models.User{
		Email:                uniqueEmail,
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Test",
		LastName:             "User",
	}

	res := as.HTML("/users").Post(u)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.Location())
}
