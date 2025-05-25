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

func (as *ActionSuite) Test_ProfileSettings_RequiresAuth() {
	res := as.HTML("/profile").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/auth/new", res.Location())
}

func (as *ActionSuite) Test_ProfileSettings_LoggedIn() {
	// Create a user first
	u := &models.User{
		Email:                "profile-test@example.com",
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Profile",
		LastName:             "Test",
	}

	verrs, err := u.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Simulate actual login like working home test
	loginData := &models.User{
		Email:    "profile-test@example.com",
		Password: "password",
	}

	// POST to login endpoint to get proper session
	loginRes := as.HTML("/auth").Post(loginData)
	as.Equal(http.StatusFound, loginRes.Code)

	// Test profile settings page
	res := as.HTML("/profile").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Profile Settings")
	as.Contains(res.Body.String(), u.Email)
}

func (as *ActionSuite) Test_ProfileUpdate_LoggedIn() {
	// Create a user first
	u := &models.User{
		Email:                "profile-update@example.com",
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Profile",
		LastName:             "Update",
	}

	verrs, err := u.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Simulate actual login like working home test
	loginData := &models.User{
		Email:    "profile-update@example.com",
		Password: "password",
	}

	// POST to login endpoint to get proper session
	loginRes := as.HTML("/auth").Post(loginData)
	as.Equal(http.StatusFound, loginRes.Code)

	// Test profile update
	updateData := &models.User{
		FirstName: "Updated",
		LastName:  "Name",
	}

	res := as.HTML("/profile").Post(updateData)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/profile", res.Location())

	// Verify the user was updated
	updatedUser := &models.User{}
	err = as.DB.Find(updatedUser, u.ID)
	as.NoError(err)
	as.Equal("Updated", updatedUser.FirstName)
	as.Equal("Name", updatedUser.LastName)
	as.Equal(u.Email, updatedUser.Email) // Email should remain unchanged
}

func (as *ActionSuite) Test_AccountSettings_RequiresAuth() {
	res := as.HTML("/account").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/auth/new", res.Location())
}

func (as *ActionSuite) Test_AccountSettings_LoggedIn() {
	// Create a user first
	u := &models.User{
		Email:                "account-test@example.com",
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Account",
		LastName:             "Test",
	}

	verrs, err := u.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Simulate actual login like working home test
	loginData := &models.User{
		Email:    "account-test@example.com",
		Password: "password",
	}

	// POST to login endpoint to get proper session
	loginRes := as.HTML("/auth").Post(loginData)
	as.Equal(http.StatusFound, loginRes.Code)

	// Test account settings page
	res := as.HTML("/account").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Account Settings")
	as.Contains(res.Body.String(), u.Email)
}
