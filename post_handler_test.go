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

func createAuthenticatedRequest(t *testing.T, testApp core.App, method, path, body string) (*core.RequestEvent, *httptest.ResponseRecorder, *core.Record) {
	collection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(collection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	token, err := user.NewAuthToken()
	require.NoError(t, err)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "pb_auth", Value: token})
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	return requestEvent, res, user
}

func TestCreatePost_ValidPost(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"title":     {"Test Blog Post"},
		"content":   {"This is the content of the test blog post."},
		"excerpt":   {"Test excerpt"},
		"published": {"on"},
	}

	requestEvent, res, user := createAuthenticatedRequest(t, testApp, http.MethodPost, "/cms/posts", formData.Encode())

	err = handleCreatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "/admin/posts", res.Header().Get("HX-Redirect"))

	collection, err := testApp.FindCollectionByNameOrId("posts")
	require.NoError(t, err)

	records, err := testApp.FindAllRecords(collection)
	require.NoError(t, err)
	require.Len(t, records, 1)

	post := records[0]
	assert.Equal(t, "Test Blog Post", post.GetString("title"))
	assert.Equal(t, "test-blog-post", post.GetString("slug"))
	assert.Equal(t, "This is the content of the test blog post.", post.GetString("content"))
	assert.Equal(t, "Test excerpt", post.GetString("excerpt"))
	assert.True(t, post.GetBool("published"))
	assert.Equal(t, user.Id, post.GetString("author"))
	assert.NotEmpty(t, post.GetDateTime("published_at"))
}

func TestCreatePost_MissingTitle(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"title":   {""},
		"content": {"This is the content."},
	}

	requestEvent, res, _ := createAuthenticatedRequest(t, testApp, http.MethodPost, "/cms/posts", formData.Encode())

	err = handleCreatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 400, res.Code)
}

func TestCreatePost_MissingContent(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"title":   {"Test Post"},
		"content": {""},
	}

	requestEvent, res, _ := createAuthenticatedRequest(t, testApp, http.MethodPost, "/cms/posts", formData.Encode())

	err = handleCreatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 400, res.Code)
}

func TestCreatePost_UnpublishedPost(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"title":   {"Draft Post"},
		"content": {"This is a draft."},
	}

	requestEvent, res, _ := createAuthenticatedRequest(t, testApp, http.MethodPost, "/cms/posts", formData.Encode())

	err = handleCreatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)

	collection, err := testApp.FindCollectionByNameOrId("posts")
	require.NoError(t, err)

	records, err := testApp.FindAllRecords(collection)
	require.NoError(t, err)
	require.Len(t, records, 1)

	post := records[0]
	assert.False(t, post.GetBool("published"))
	assert.Empty(t, post.GetDateTime("published_at").String())
}

func TestUpdatePost_ValidUpdate(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	collection, err := testApp.FindCollectionByNameOrId("posts")
	require.NoError(t, err)

	userCollection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(userCollection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	post := core.NewRecord(collection)
	post.Set("title", "Original Title")
	post.Set("slug", "original-title")
	post.Set("content", "Original content")
	post.Set("published", false)
	post.Set("author", user.Id)
	err = testApp.Save(post)
	require.NoError(t, err)

	formData := url.Values{
		"title":     {"Updated Title"},
		"content":   {"Updated content"},
		"excerpt":   {"Updated excerpt"},
		"published": {"on"},
	}

	token, err := user.NewAuthToken()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/cms/posts/"+post.Id, strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "pb_auth", Value: token})
	req.SetPathValue("id", post.Id)
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleUpdatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "/admin/posts", res.Header().Get("HX-Redirect"))

	updatedPost, err := testApp.FindRecordById("posts", post.Id)
	require.NoError(t, err)

	assert.Equal(t, "Updated Title", updatedPost.GetString("title"))
	assert.Equal(t, "updated-title", updatedPost.GetString("slug"))
	assert.Equal(t, "Updated content", updatedPost.GetString("content"))
	assert.Equal(t, "Updated excerpt", updatedPost.GetString("excerpt"))
	assert.True(t, updatedPost.GetBool("published"))
	assert.NotEmpty(t, updatedPost.GetDateTime("published_at"))
}

func TestUpdatePost_InvalidPostId(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"title":   {"Updated Title"},
		"content": {"Updated content"},
	}

	userCollection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(userCollection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	token, err := user.NewAuthToken()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/cms/posts/invalid-id", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "pb_auth", Value: token})
	req.SetPathValue("id", "invalid-id")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleUpdatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 404, res.Code)
}

func TestDeletePost_ValidDeletion(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	collection, err := testApp.FindCollectionByNameOrId("posts")
	require.NoError(t, err)

	userCollection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(userCollection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	post := core.NewRecord(collection)
	post.Set("title", "Post to Delete")
	post.Set("slug", "post-to-delete")
	post.Set("content", "This will be deleted")
	post.Set("author", user.Id)
	err = testApp.Save(post)
	require.NoError(t, err)

	token, err := user.NewAuthToken()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/cms/posts/"+post.Id, nil)
	req.AddCookie(&http.Cookie{Name: "pb_auth", Value: token})
	req.SetPathValue("id", post.Id)
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleDeletePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "/admin/posts", res.Header().Get("HX-Redirect"))

	_, err = testApp.FindRecordById("posts", post.Id)
	assert.Error(t, err)
}

func TestDeletePost_InvalidPostId(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	userCollection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(userCollection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	token, err := user.NewAuthToken()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/cms/posts/invalid-id", nil)
	req.AddCookie(&http.Cookie{Name: "pb_auth", Value: token})
	req.SetPathValue("id", "invalid-id")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleDeletePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 404, res.Code)
}

func TestBlogPost_MarkdownRendering(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	collection, err := testApp.FindCollectionByNameOrId("posts")
	require.NoError(t, err)

	userCollection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(userCollection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	markdownContent := `# Heading 1
## Heading 2
**Bold text** and *italic text*

- List item 1
- List item 2

` + "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" + `

> This is a blockquote

| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |`

	post := core.NewRecord(collection)
	post.Set("title", "Markdown Test Post")
	post.Set("slug", "markdown-test")
	post.Set("content", markdownContent)
	post.Set("published", true)
	post.Set("author", user.Id)
	err = testApp.Save(post)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/blog/markdown-test", nil)
	req.SetPathValue("slug", "markdown-test")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleBlogPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)

	body := res.Body.String()
	assert.Contains(t, body, "<h1")
	assert.Contains(t, body, "<h2")
	assert.Contains(t, body, "<strong>Bold text</strong>")
	assert.Contains(t, body, "<em>italic text</em>")
	assert.Contains(t, body, "<li>List item 1</li>")
	assert.Contains(t, body, "<li>List item 2</li>")
	assert.Contains(t, body, "<pre>")
	assert.Contains(t, body, "<code")
	assert.Contains(t, body, "func main()")
	assert.Contains(t, body, "<blockquote>")
	assert.Contains(t, body, "<table>")
	assert.Contains(t, body, "<th>Header 1</th>")
	assert.Contains(t, body, "<td>Cell 1</td>")

	assert.NotContains(t, body, "# Heading 1")
	assert.NotContains(t, body, "## Heading 2")
	assert.NotContains(t, body, "**Bold text**")
	assert.NotContains(t, body, "```go")
}

func TestBlogPost_MarkdownSanitization(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	collection, err := testApp.FindCollectionByNameOrId("posts")
	require.NoError(t, err)

	userCollection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(userCollection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	maliciousMarkdown := `# Safe Heading
<script>alert('XSS')</script>
**Bold text**
<img src=x onerror="alert('XSS')">
[Link](javascript:alert('XSS'))`

	post := core.NewRecord(collection)
	post.Set("title", "Security Test Post")
	post.Set("slug", "security-test")
	post.Set("content", maliciousMarkdown)
	post.Set("published", true)
	post.Set("author", user.Id)
	err = testApp.Save(post)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/blog/security-test", nil)
	req.SetPathValue("slug", "security-test")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleBlogPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)

	body := res.Body.String()
	assert.NotContains(t, body, "<script>")
	assert.NotContains(t, body, "alert('XSS')")
	assert.NotContains(t, body, "onerror=")
	assert.NotContains(t, body, "javascript:")
	assert.Contains(t, body, "<h1")
	assert.Contains(t, body, "Safe Heading")
	assert.Contains(t, body, "<strong>Bold text</strong>")
}

func TestBlogPost_AdminCanViewDraft(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	collection, err := testApp.FindCollectionByNameOrId("posts")
	require.NoError(t, err)

	userCollection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(userCollection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	token, err := user.NewAuthToken()
	require.NoError(t, err)

	draftPost := core.NewRecord(collection)
	draftPost.Set("title", "Draft Post")
	draftPost.Set("slug", "draft-post")
	draftPost.Set("content", "# Draft Content\nThis is a draft.")
	draftPost.Set("published", false)
	draftPost.Set("author", user.Id)
	err = testApp.Save(draftPost)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/blog/draft-post", nil)
	req.SetPathValue("slug", "draft-post")
	req.AddCookie(&http.Cookie{Name: "pb_auth", Value: token})
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleBlogPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, "Draft Post")
	assert.Contains(t, body, "<h1")
	assert.Contains(t, body, "Draft Content")
}

func TestBlogPost_NonAdminCannotViewDraft(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	collection, err := testApp.FindCollectionByNameOrId("posts")
	require.NoError(t, err)

	userCollection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(userCollection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	draftPost := core.NewRecord(collection)
	draftPost.Set("title", "Draft Post")
	draftPost.Set("slug", "draft-post-private")
	draftPost.Set("content", "# Draft Content\nThis is a draft.")
	draftPost.Set("published", false)
	draftPost.Set("author", user.Id)
	err = testApp.Save(draftPost)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/blog/draft-post-private", nil)
	req.SetPathValue("slug", "draft-post-private")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleBlogPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 404, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, "Post Not Found")
}

func TestEditPost_LoadsCorrectly(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	userCollection, err := testApp.FindCollectionByNameOrId("users")
	require.NoError(t, err)

	user := core.NewRecord(userCollection)
	user.Set("email", "admin@example.com")
	user.Set("role", "admin")
	user.SetPassword("password123")
	err = testApp.Save(user)
	require.NoError(t, err)

	token, err := user.NewAuthToken()
	require.NoError(t, err)

	collection, err := testApp.FindCollectionByNameOrId("posts")
	require.NoError(t, err)

	post := core.NewRecord(collection)
	post.Set("title", "Test Post for Editing")
	post.Set("slug", "test-post-edit")
	post.Set("content", "Original content")
	post.Set("excerpt", "Original excerpt")
	post.Set("published", true)
	post.Set("author", user.Id)
	err = testApp.Save(post)
	require.NoError(t, err)

	t.Logf("Created post with ID: %s", post.Id)

	req := httptest.NewRequest(http.MethodGet, "/cms/posts/"+post.Id+"/edit", nil)
	req.AddCookie(&http.Cookie{Name: "pb_auth", Value: token})
	req.SetPathValue("id", post.Id)
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleEditPost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, "Test Post for Editing")
	assert.Contains(t, body, "Original content")
	assert.Contains(t, body, post.Id)
}
