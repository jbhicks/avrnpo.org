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

	"avrnpo.org/services"
)

func TestContactHandler_ValidSubmission(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	emailService = &services.EmailService{EmailEnabled: false}

	formData := url.Values{
		"name":    {"John Doe"},
		"email":   {"john@example.com"},
		"phone":   {"555-1234"},
		"message": {"This is a test message with more than 10 characters"},
	}

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleContactPost(requestEvent)
	require.NoError(t, err)

	collection, err := testApp.FindCollectionByNameOrId("contact_submissions")
	require.NoError(t, err)

	records, err := testApp.FindAllRecords(collection)
	require.NoError(t, err)
	require.Len(t, records, 1)

	record := records[0]
	assert.Equal(t, "John Doe", record.GetString("name"))
	assert.Equal(t, "john@example.com", record.GetString("email"))
	assert.Equal(t, "555-1234", record.GetString("phone"))
	assert.Equal(t, "This is a test message with more than 10 characters", record.GetString("message"))
	assert.Equal(t, "new", record.GetString("status"))
}

func TestContactHandler_MissingName(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	emailService = &services.EmailService{EmailEnabled: false}

	formData := url.Values{
		"name":    {""},
		"email":   {"john@example.com"},
		"phone":   {"555-1234"},
		"message": {"This is a test message"},
	}

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleContactPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 400, res.Code)
}

func TestContactHandler_InvalidEmail(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	emailService = &services.EmailService{EmailEnabled: false}

	formData := url.Values{
		"name":    {"John Doe"},
		"email":   {"invalid-email"},
		"phone":   {"555-1234"},
		"message": {"This is a test message"},
	}

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleContactPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 400, res.Code)
}

func TestContactHandler_MessageTooShort(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	emailService = &services.EmailService{EmailEnabled: false}

	formData := url.Values{
		"name":    {"John Doe"},
		"email":   {"john@example.com"},
		"phone":   {"555-1234"},
		"message": {"Short"},
	}

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleContactPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 400, res.Code)
}

func TestContactHandler_PhoneOptional(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	emailService = &services.EmailService{EmailEnabled: false}

	formData := url.Values{
		"name":    {"John Doe"},
		"email":   {"john@example.com"},
		"phone":   {""},
		"message": {"This is a test message with more than 10 characters"},
	}

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleContactPost(requestEvent)
	require.NoError(t, err)

	collection, err := testApp.FindCollectionByNameOrId("contact_submissions")
	require.NoError(t, err)

	records, err := testApp.FindAllRecords(collection)
	require.NoError(t, err)
	require.Len(t, records, 1)

	record := records[0]
	assert.Equal(t, "", record.GetString("phone"))
}
