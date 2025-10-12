package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginHandler_ValidCredentials(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	collection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("email", "testuser@example.com")
	record.Set("role", "admin")
	record.SetPassword("password123")
	err = testApp.Save(record)
	require.NoError(t, err)

	formData := url.Values{
		"email":    {"testuser@example.com"},
		"password": {"password123"},
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleLoginPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "/cms/posts", res.Header().Get("HX-Redirect"))

	cookies := res.Header().Get("Set-Cookie")
	assert.Contains(t, cookies, "pb_auth=")
	assert.Contains(t, cookies, "HttpOnly")
	assert.Contains(t, cookies, "SameSite=Lax")
}

func TestLoginHandler_InvalidEmail(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"email":    {"nonexistent@example.com"},
		"password": {"password123"},
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleLoginPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)
	assert.Contains(t, res.Body.String(), "Invalid email or password")
}

func TestLoginHandler_InvalidPassword(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	collection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("email", "testuser@example.com")
	record.Set("role", "admin")
	record.SetPassword("correctpassword")
	err = testApp.Save(record)
	require.NoError(t, err)

	formData := url.Values{
		"email":    {"testuser@example.com"},
		"password": {"wrongpassword"},
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleLoginPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)
	assert.Contains(t, res.Body.String(), "Invalid email or password")
}
