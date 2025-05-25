package actions

import (
	"net/http"

	"my_go_saas_template/models"
)

func (as *ActionSuite) Test_HomeHandler() {
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Build Your SaaS")
	as.Contains(res.Body.String(), "The Right Way")
}

func (as *ActionSuite) Test_HomeHandler_LoggedIn() {
	u := &models.User{
		Email:                "mark@example.com",
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Mark",
		LastName:             "Smith",
	}
	verrs, err := u.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())
	as.Session.Set("current_user_id", u.ID)

	// Test that logged in users still see the landing page with dashboard link
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Go to Dashboard")

	// Test that the dashboard is accessible
	res = as.HTML("/dashboard").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Welcome to Your SaaS Dashboard")
}
