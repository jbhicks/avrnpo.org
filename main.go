package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"avrnpo.org/middleware"
	_ "avrnpo.org/migrations"
	"avrnpo.org/services"
	"avrnpo.org/templates"
)

var helcimClient services.HelcimAPI
var emailService *services.EmailService

var (
	loginRateLimiter   *middleware.RateLimiter
	contactRateLimiter *middleware.RateLimiter
	donateRateLimiter  *middleware.RateLimiter
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	app := pocketbase.New()

	helcimClient = services.NewHelcimClient()
	emailService = services.NewEmailService()

	loginRateLimiter = middleware.NewRateLimiter(5, time.Minute)
	contactRateLimiter = middleware.NewRateLimiter(3, time.Minute)
	donateRateLimiter = middleware.NewRateLimiter(10, time.Minute)

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: os.Getenv("PB_AUTOMIGRATE") == "1",
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		adminEmail := os.Getenv("PB_ADMIN_EMAIL")
		adminPassword := os.Getenv("PB_ADMIN_PASSWORD")
		if adminEmail != "" && adminPassword != "" {
			collection, err := se.App.FindCollectionByNameOrId("users")
			if err != nil {
				log.Printf("Users collection not found, skipping admin setup: %v", err)
				return se.Next()
			}

			existingUser, err := se.App.FindFirstRecordByFilter("users", "email = {:email}", map[string]any{"email": adminEmail})
			if err == nil {
				if existingUser.GetString("role") != "admin" {
					existingUser.Set("role", "admin")
					if err := se.App.Save(existingUser); err != nil {
						log.Printf("Failed to update user role to admin: %v", err)
					} else {
						log.Printf("Updated user %s to admin role", adminEmail)
					}
				} else {
					log.Printf("Admin user already exists: %s", adminEmail)
				}
			} else {
				admin := core.NewRecord(collection)
				admin.Set("email", adminEmail)
				admin.Set("username", adminEmail)
				admin.Set("role", "admin")
				admin.SetPassword(adminPassword)

				if err := se.App.Save(admin); err != nil {
					log.Printf("Failed to create admin user: %v", err)
				} else {
					log.Printf("Created new admin user: %s", adminEmail)
				}
			}
		}

		return se.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		log.Printf("Registering custom routes...")

		se.Router.GET("/auth/login", handleLoginPage)
		se.Router.POST("/auth/login", loginRateLimiter.RequestEventMiddleware(middleware.CSRFProtection(handleLoginPost)))
		se.Router.POST("/auth/logout", middleware.CSRFProtection(handleLogout))

		se.Router.GET("/cms/posts/new", handleNewPost)
		se.Router.POST("/cms/posts", middleware.CSRFProtection(handleCreatePost))
		se.Router.GET("/cms/posts/{id}/edit", handleEditPost)
		se.Router.PUT("/cms/posts/{id}", middleware.CSRFProtection(handleUpdatePost))
		se.Router.DELETE("/cms/posts/{id}", middleware.CSRFProtection(handleDeletePost))
		se.Router.GET("/cms/posts", handleAdminPostList)
		log.Printf("Registered /cms/posts routes")

		se.Router.GET("/blog", handleBlogList)
		log.Printf("Registered /blog route")
		se.Router.GET("/blog/{slug}", handleBlogPost)
		log.Printf("Registered /blog/{slug} route")

		se.Router.GET("/donate/success", handleDonateSuccess)
		se.Router.GET("/donate/failed", handleDonateFailed)
		se.Router.GET("/donate/fragment", handleDonateFragment)
		se.Router.GET("/donate", handleDonate)
		se.Router.POST("/donate", donateRateLimiter.RequestEventMiddleware(middleware.CSRFProtection(handleDonatePost)))
		se.Router.POST("/api/donations/process", middleware.CSRFProtection(handleProcessPayment))
		se.Router.GET("/api/nav-state", handleNavState)
		se.Router.GET("/contact", handleContact)
		se.Router.GET("/contact/fragment", handleContactFragment)
		se.Router.POST("/contact", contactRateLimiter.RequestEventMiddleware(middleware.CSRFProtection(handleContactPost)))
		se.Router.GET("/team", handleTeam)
		se.Router.GET("/projects", handleProjects)
		se.Router.GET("/about", handleAbout)

		se.Router.GET("/", handleHome)
		se.Router.GET("/assets/{path...}", apis.Static(os.DirFS("./pb_public/assets"), false))
		log.Printf("Registered static file handler at /assets/{path...}")

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func handleHome(e *core.RequestEvent) error {
	log.Printf("[HOME] Handler called for URL: %s", e.Request.URL.Path)
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("Error generating CSRF token: %v", err)
		return e.HTML(500, "Internal server error")
	}

	posts, err := e.App.FindRecordsByFilter(
		"posts",
		"published = true",
		"-published_at",
		3,
		0,
	)

	templatePosts := make([]templates.Post, 0, len(posts))
	if err == nil {
		for _, post := range posts {
			templatePosts = append(templatePosts, templates.Post{
				Slug:        post.GetString("slug"),
				Title:       post.GetString("title"),
				Excerpt:     post.GetString("excerpt"),
				PublishedAt: post.GetDateTime("published_at").Time().Format("January 2, 2006"),
			})
		}
	}

	return templates.HomePage(templatePosts, csrfToken).Render(e.Request.Context(), e.Response)
}

func handleBlogList(e *core.RequestEvent) error {
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("Error generating CSRF token: %v", err)
		csrfToken = ""
	}

	posts, err := e.App.FindRecordsByFilter(
		"posts",
		"published = true",
		"-published_at",
		20,
		0,
	)
	if err != nil {
		log.Printf("Error fetching blog posts: %v", err)
		return e.JSON(500, map[string]string{"error": "Failed to fetch blog posts"})
	}

	templatePosts := make([]templates.Post, 0, len(posts))
	for _, post := range posts {
		templatePosts = append(templatePosts, templates.Post{
			Slug:        post.GetString("slug"),
			Title:       post.GetString("title"),
			Excerpt:     post.GetString("excerpt"),
			PublishedAt: post.GetDateTime("published_at").Time().Format("January 2, 2006"),
		})
	}

	return templates.BlogIndexPage(templatePosts, csrfToken).Render(e.Request.Context(), e.Response)
}

func handleBlogPost(e *core.RequestEvent) error {
	slug := e.Request.PathValue("slug")
	log.Printf("[BLOG POST] Handler called for slug: %s", slug)
	log.Printf("[BLOG POST] Full URL: %s", e.Request.URL.Path)

	// Check if user is admin
	_, err := isAdmin(e)
	isAdminUser := err == nil

	// Build filter based on admin status
	filter := "slug = {:slug}"
	if !isAdminUser {
		filter += " && published = true"
	}

	post, err := e.App.FindFirstRecordByFilter(
		"posts",
		filter,
		map[string]any{"slug": slug},
	)
	if err != nil {
		return e.HTML(404, `
			<!DOCTYPE html>
			<html>
			<head><title>Post Not Found - AVR NPO</title></head>
			<body>
				<h1>Post Not Found</h1>
				<p>The blog post you're looking for doesn't exist or hasn't been published yet.</p>
				<a href="/blog">Back to Updates</a>
			</body>
			</html>
		`)
	}

	templatePost := templates.Post{
		Slug:        slug,
		Title:       post.GetString("title"),
		Content:     services.SanitizeContent(post.GetString("content")),
		PublishedAt: post.GetDateTime("published_at").Time().Format("January 2, 2006"),
	}

	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		csrfToken = ""
	}

	return templates.BlogPostPage(templatePost, csrfToken).Render(e.Request.Context(), e.Response)
}

func handleDonate(e *core.RequestEvent) error {
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("Error generating CSRF token: %v", err)
		return e.HTML(500, "Internal server error")
	}
	log.Printf("[DONATE] GET /donate - Generated CSRF token: %s (length: %d)", csrfToken[:10], len(csrfToken))
	return templates.DonatePage(csrfToken).Render(e.Request.Context(), e.Response)
}

func handleDonateFragment(e *core.RequestEvent) error {
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("Error generating CSRF token for donate fragment: %v", err)
		return e.HTML(500, "Internal server error")
	}
	return templates.DonateFragment(csrfToken).Render(e.Request.Context(), e.Response)
}

func handleDonatePost(e *core.RequestEvent) error {
	donationType := e.Request.FormValue("donation_type")
	amountStr := e.Request.FormValue("amount")
	name := e.Request.FormValue("name")
	email := e.Request.FormValue("email")
	addressLine1 := e.Request.FormValue("address_line1")
	city := e.Request.FormValue("city")
	province := e.Request.FormValue("province")
	postalCode := e.Request.FormValue("postal_code")
	country := e.Request.FormValue("country")

	log.Printf("[DONATE] Form values - province: '%s', postal_code: '%s', country: '%s'", province, postalCode, country)

	if err := middleware.ValidateDonationType(donationType); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return e.JSON(400, map[string]string{"error": "Invalid amount format"})
	}

	if err := middleware.ValidateAmount(amount); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}

	if err := middleware.ValidateRequired("name", name); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}

	if err := middleware.ValidateEmail(email); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}

	collection, err := e.App.FindCollectionByNameOrId("donations")
	if err != nil {
		return e.JSON(500, map[string]string{"error": "Collection not found"})
	}

	record := core.NewRecord(collection)
	record.Set("amount", amount)
	record.Set("currency", "USD")
	record.Set("donation_type", donationType)
	record.Set("status", "pending")
	record.Set("donor_name", name)
	record.Set("donor_email", email)
	record.Set("address_line1", addressLine1)
	record.Set("city", city)
	record.Set("province", province)
	record.Set("postal_code", postalCode)
	record.Set("country", country)

	if err := e.App.Save(record); err != nil {
		e.App.Logger().Error("Failed to save donation", "error", err)
		return e.JSON(500, map[string]string{"error": "Failed to save donation"})
	}

	initReq := services.InitializeRequest{
		PaymentType: "purchase",
		Amount:      amount,
		Currency:    "USD",
		CustomerRequest: &services.CustomerRequest{
			ContactName: name,
			Email:       email,
			BillingAddress: services.BillingAddress{
				Name:       name,
				Street1:    addressLine1,
				City:       city,
				Province:   province,
				Country:    country,
				PostalCode: postalCode,
			},
		},
	}

	log.Printf("[DONATE] Submitting to Helcim - Name: %s, Email: %s, Address: %s, City: %s, Province: %s, Country: %s, Postal: %s, Amount: %.2f",
		name, email, addressLine1, city, province, country, postalCode, amount)

	initResp, err := helcimClient.Initialize(initReq)
	if err != nil {
		record.Set("status", "failed")
		record.Set("error_message", err.Error())
		e.App.Save(record)
		return e.JSON(500, map[string]string{"error": "Failed to initialize payment"})
	}

	record.Set("checkout_token", initResp.CheckoutToken)
	record.Set("secret_token", initResp.SecretToken)
	if err := e.App.Save(record); err != nil {
		return e.JSON(500, map[string]string{"error": "Failed to update donation"})
	}

	newCsrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("[DONATE] Error generating new CSRF token: %v", err)
		return e.JSON(500, map[string]string{"error": "Failed to generate security token"})
	}

	return e.JSON(200, map[string]any{
		"donationId":    record.Id,
		"checkoutToken": initResp.CheckoutToken,
		"secretToken":   initResp.SecretToken,
		"csrfToken":     newCsrfToken,
	})
}

func handleProcessPayment(e *core.RequestEvent) error {
	log.Printf("[PAYMENT_PROCESS] Received payment processing request from %s", e.Request.RemoteAddr)

	var req struct {
		DonationID    string  `json:"donationId"`
		CustomerCode  string  `json:"customerCode"`
		CardToken     string  `json:"cardToken"`
		Amount        float64 `json:"amount"`
		TransactionID string  `json:"transactionId"`
	}

	log.Printf("[PAYMENT_PROCESS] Binding request body...")
	if err := e.BindBody(&req); err != nil {
		log.Printf("[PAYMENT_PROCESS] Failed to bind request body: %v", err)
		return e.JSON(400, map[string]string{"error": "Invalid request"})
	}

	log.Printf("[PAYMENT_PROCESS] Request data: donationId=%s, customerCode=%s, amount=%.2f",
		req.DonationID, req.CustomerCode, req.Amount)

	collection, err := e.App.FindCollectionByNameOrId("donations")
	if err != nil {
		return e.JSON(500, map[string]string{"error": "Collection not found"})
	}

	record, err := e.App.FindRecordById(collection, req.DonationID)
	if err != nil {
		return e.JSON(404, map[string]string{"error": "Donation not found"})
	}

	donationType := record.GetString("donation_type")
	amount := record.GetFloat("amount")

	if donationType == "one-time" {
		// HelcimPay.js already processed the payment - just record the transaction
		log.Printf("[PAYMENT] Recording HelcimPay transaction %s for donation %s", req.TransactionID, req.DonationID)

		record.Set("helcim_transaction_id", req.TransactionID)
		record.Set("customer_id", req.CustomerCode)
		record.Set("status", "completed")

		log.Printf("[PAYMENT] Saving donation record %s", req.DonationID)
		if err := e.App.Save(record); err != nil {
			log.Printf("[PAYMENT] Failed to save record for donation %s: %v", req.DonationID, err)
			return e.JSON(500, map[string]string{"error": "Failed to save record"})
		}
		log.Printf("[PAYMENT] Successfully saved donation record %s", req.DonationID)

		go sendDonationReceipt(record, req.TransactionID, "")

		return e.JSON(200, map[string]any{
			"status":         "success",
			"transaction_id": req.TransactionID,
			"message":        "Thank you for your donation!",
		})

	} else if donationType == "monthly" {
		planName := fmt.Sprintf("Monthly $%.2f Donation", amount)
		plan, err := helcimClient.CreatePaymentPlan(amount, planName)
		if err != nil {
			record.Set("status", "failed")
			record.Set("error_message", err.Error())
			e.App.Save(record)
			log.Printf("Failed to create payment plan: %v", err)
			return e.JSON(500, map[string]string{"error": "Failed to create payment plan"})
		}

		subscriptionReq := services.SubscriptionRequest{
			CustomerID:    req.CustomerCode,
			PaymentPlanID: plan.ID,
			Amount:        amount,
			PaymentMethod: "card",
		}

		subscriptionResp, err := helcimClient.CreateSubscription(subscriptionReq)
		if err != nil {
			record.Set("status", "failed")
			record.Set("error_message", err.Error())
			e.App.Save(record)
			log.Printf("Failed to create subscription: %v", err)
			return e.JSON(500, map[string]string{"error": "Failed to create subscription"})
		}

		record.Set("subscription_id", subscriptionResp.ID)
		record.Set("payment_plan_id", plan.ID)
		record.Set("customer_id", req.CustomerCode)
		record.Set("status", "completed")
		record.Set("subscription_status", "active")
		record.Set("next_billing_date", subscriptionResp.NextBillingDate)

		if err := e.App.Save(record); err != nil {
			return e.JSON(500, map[string]string{"error": "Failed to save record"})
		}

		go sendDonationReceipt(record, "", fmt.Sprintf("%d", subscriptionResp.ID))

		return e.JSON(200, map[string]any{
			"status":          "success",
			"subscription_id": subscriptionResp.ID,
			"message":         "Thank you for your recurring donation!",
		})
	}

	return e.JSON(400, map[string]string{"error": "Invalid donation type"})
}

func sendDonationReceipt(record *core.Record, transactionID string, subscriptionID string) {
	donorEmail := record.GetString("donor_email")
	donorName := record.GetString("donor_name")
	amount := record.GetFloat("amount")
	donationType := record.GetString("donation_type")

	receiptData := services.DonationReceiptData{
		DonorName:           donorName,
		DonationAmount:      amount,
		DonationType:        donationType,
		SubscriptionID:      subscriptionID,
		TransactionID:       transactionID,
		DonationDate:        time.Now(),
		TaxDeductibleAmount: amount,
		OrganizationEIN:     os.Getenv("ORGANIZATION_EIN"),
		OrganizationName:    "American Veterans Rebuilding",
		OrganizationAddress: os.Getenv("ORGANIZATION_ADDRESS"),
		DonorAddressLine1:   record.GetString("address_line1"),
		DonorCity:           record.GetString("city"),
		DonorState:          record.GetString("province"),
		DonorZip:            record.GetString("postal_code"),
	}

	if donationType == "monthly" {
		nextBillingDate := record.GetDateTime("next_billing_date")
		if !nextBillingDate.IsZero() {
			t := nextBillingDate.Time()
			receiptData.NextBillingDate = &t
		}
	}

	if err := emailService.SendDonationReceipt(donorEmail, receiptData); err != nil {
		log.Printf("[ERROR] Failed to send donation receipt to %s: %v", donorEmail, err)
	} else {
		log.Printf("[INFO] Donation receipt sent successfully to %s", donorEmail)
	}
}

func handleDonateSuccess(e *core.RequestEvent) error {
	return templates.DonationSuccessPage().Render(e.Request.Context(), e.Response)
}

func handleDonateFailed(e *core.RequestEvent) error {
	return templates.DonationFailedPage().Render(e.Request.Context(), e.Response)
}

func handleContact(e *core.RequestEvent) error {
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("Error generating CSRF token: %v", err)
		return e.HTML(500, "Internal server error")
	}
	return templates.ContactPage(csrfToken).Render(e.Request.Context(), e.Response)
}

func handleContactFragment(e *core.RequestEvent) error {
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("Error generating CSRF token for fragment: %v", err)
		return e.HTML(500, "Internal server error")
	}
	return templates.ContactFragment(csrfToken).Render(e.Request.Context(), e.Response)
}

func handleContactPost(e *core.RequestEvent) error {
	name := e.Request.FormValue("name")
	email := e.Request.FormValue("email")
	phone := e.Request.FormValue("phone")
	message := e.Request.FormValue("message")

	if err := middleware.ValidateRequired("name", name); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := middleware.ValidateEmail(email); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := middleware.ValidateRequired("message", message); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := middleware.ValidateLength("message", message, 10, 5000); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}

	collection, err := e.App.FindCollectionByNameOrId("contact_submissions")
	if err != nil {
		return e.JSON(500, map[string]string{"error": "Collection not found"})
	}

	record := core.NewRecord(collection)
	record.Set("name", name)
	record.Set("email", email)
	record.Set("phone", phone)
	record.Set("message", message)
	record.Set("status", "new")

	if err := e.App.Save(record); err != nil {
		return e.JSON(400, map[string]string{"error": "Failed to save contact submission"})
	}

	go sendContactNotification(name, email, message)

	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		csrfToken = ""
	}

	return templates.ContactSuccessPage(csrfToken).Render(e.Request.Context(), e.Response)
}

func sendContactNotification(name, email, message string) {
	contactData := services.ContactFormData{
		Name:           name,
		Email:          email,
		Subject:        "New Contact Form Submission",
		Message:        message,
		SubmissionDate: time.Now(),
	}

	if err := emailService.SendContactNotification(contactData); err != nil {
		log.Printf("[ERROR] Failed to send contact notification: %v", err)
	} else {
		log.Printf("[INFO] Contact notification sent successfully for %s", email)
	}
}

func handleAbout(e *core.RequestEvent) error {
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		csrfToken = ""
	}
	return templates.AboutPage(csrfToken).Render(e.Request.Context(), e.Response)
}

func handleTeam(e *core.RequestEvent) error {
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		csrfToken = ""
	}
	return templates.TeamPage(csrfToken).Render(e.Request.Context(), e.Response)
}

func handleProjects(e *core.RequestEvent) error {
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		csrfToken = ""
	}
	return templates.ProjectsPage(csrfToken).Render(e.Request.Context(), e.Response)
}

func handleNavState(e *core.RequestEvent) error {
	e.Response.Header().Set("Content-Type", "text/html")
	return templates.NavigationScrolled().Render(e.Request.Context(), e.Response)
}

func handleLoginPage(e *core.RequestEvent) error {
	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("Error generating CSRF token: %v", err)
		return e.HTML(500, "Internal server error")
	}

	errorMsg := e.Request.URL.Query().Get("error")
	return templates.LoginPage(errorMsg, csrfToken).Render(e.Request.Context(), e.Response)
}

func handleLoginPost(e *core.RequestEvent) error {
	email := e.Request.FormValue("email")
	password := e.Request.FormValue("password")

	log.Printf("[LOGIN] Attempt for email: %s", email)

	user, err := e.App.FindAuthRecordByEmail("users", email)
	if err != nil {
		log.Printf("[LOGIN] Failed to find user %s: %v", email, err)
		csrfToken, _ := middleware.GetCSRFToken(e)
		return templates.LoginPage("Invalid email or password", csrfToken).Render(e.Request.Context(), e.Response)
	}

	if !user.ValidatePassword(password) {
		log.Printf("[LOGIN] Invalid password for user: %s", email)
		csrfToken, _ := middleware.GetCSRFToken(e)
		return templates.LoginPage("Invalid email or password", csrfToken).Render(e.Request.Context(), e.Response)
	}

	token, err := user.NewAuthToken()
	if err != nil {
		log.Printf("Error creating auth token: %v", err)
		csrfToken, _ := middleware.GetCSRFToken(e)
		return templates.LoginPage("Login failed. Please try again.", csrfToken).Render(e.Request.Context(), e.Response)
	}

	isSecure := os.Getenv("APP_ENV") == "production" || e.Request.TLS != nil
	secureFlagStr := ""
	if isSecure {
		secureFlagStr = "; Secure"
	}

	cookieStr := fmt.Sprintf("pb_auth=%s; Path=/; HttpOnly; SameSite=Lax; Max-Age=%d%s", token, 60*60*24*7, secureFlagStr)
	e.Response.Header().Set("Set-Cookie", cookieStr)
	e.Response.Header().Set("HX-Redirect", "/cms/posts")
	return e.NoContent(200)
}

func handleLogout(e *core.RequestEvent) error {
	e.Response.Header().Set("Set-Cookie", "pb_auth=; Path=/; HttpOnly; Max-Age=-1")
	e.Response.Header().Set("HX-Redirect", "/")
	return e.NoContent(200)
}

func isAdmin(e *core.RequestEvent) (*core.Record, error) {
	cookie, err := e.Request.Cookie("pb_auth")
	if err != nil {
		log.Printf("isAdmin: no cookie: %v", err)
		return nil, fmt.Errorf("not authenticated")
	}

	user, err := e.App.FindAuthRecordByToken(cookie.Value, core.TokenTypeAuth)
	if err != nil {
		log.Printf("isAdmin: invalid token: %v", err)
		return nil, fmt.Errorf("invalid token")
	}

	role := user.GetString("role")
	log.Printf("isAdmin: user %s has role: %s", user.GetString("email"), role)
	if role != "admin" {
		return nil, fmt.Errorf("not authorized")
	}

	return user, nil
}

func handleNewPost(e *core.RequestEvent) error {
	if _, err := isAdmin(e); err != nil {
		return e.Redirect(http.StatusSeeOther, "/auth/login")
	}

	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("Error generating CSRF token: %v", err)
		return e.HTML(500, "Internal server error")
	}

	return templates.PostForm(nil, csrfToken).Render(e.Request.Context(), e.Response)
}

func handleCreatePost(e *core.RequestEvent) error {
	user, err := isAdmin(e)
	if err != nil {
		return e.Redirect(http.StatusSeeOther, "/auth/login")
	}

	collection, err := e.App.FindCollectionByNameOrId("posts")
	if err != nil {
		return e.JSON(500, map[string]string{"error": "Collection not found"})
	}

	record := core.NewRecord(collection)
	title := e.Request.FormValue("title")
	content := e.Request.FormValue("content")
	excerpt := e.Request.FormValue("excerpt")
	published := e.Request.FormValue("published") == "on"

	if err := middleware.ValidateRequired("title", title); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := middleware.ValidateRequired("content", content); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}

	slug := middleware.SanitizeSlug(title)

	record.Set("title", title)
	record.Set("slug", slug)
	record.Set("content", content)
	record.Set("excerpt", excerpt)
	record.Set("published", published)
	record.Set("author", user.Id)
	if published {
		record.Set("published_at", time.Now())
	}

	if err := e.App.Save(record); err != nil {
		return e.JSON(400, map[string]string{"error": "Failed to save post"})
	}

	e.Response.Header().Set("HX-Redirect", "/admin/posts")
	return e.NoContent(200)
}

func handleEditPost(e *core.RequestEvent) error {
	log.Printf("[EDIT POST] Handler called, URL: %s", e.Request.URL.Path)

	if _, err := isAdmin(e); err != nil {
		log.Printf("[EDIT POST] Admin check failed, redirecting to login")
		return e.Redirect(http.StatusSeeOther, "/auth/login")
	}

	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		log.Printf("Error generating CSRF token: %v", err)
		return e.HTML(500, "Internal server error")
	}

	postId := e.Request.PathValue("id")
	log.Printf("[EDIT POST] Post ID: %s", postId)
	post, err := e.App.FindRecordById("posts", postId)
	if err != nil {
		return e.HTML(404, "Post not found")
	}

	postData := &templates.PostFormData{
		Id:        post.Id,
		Title:     post.GetString("title"),
		Content:   post.GetString("content"),
		Excerpt:   post.GetString("excerpt"),
		Published: post.GetBool("published"),
	}

	return templates.PostForm(postData, csrfToken).Render(e.Request.Context(), e.Response)
}

func handleUpdatePost(e *core.RequestEvent) error {
	if _, err := isAdmin(e); err != nil {
		return e.Redirect(http.StatusSeeOther, "/auth/login")
	}

	postId := e.Request.PathValue("id")
	post, err := e.App.FindRecordById("posts", postId)
	if err != nil {
		return e.JSON(404, map[string]string{"error": "Post not found"})
	}

	title := e.Request.FormValue("title")
	content := e.Request.FormValue("content")
	excerpt := e.Request.FormValue("excerpt")
	published := e.Request.FormValue("published") == "on"

	if err := middleware.ValidateRequired("title", title); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := middleware.ValidateRequired("content", content); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}

	slug := middleware.SanitizeSlug(title)

	post.Set("title", title)
	post.Set("slug", slug)
	post.Set("content", content)
	post.Set("excerpt", excerpt)

	wasPublished := post.GetBool("published")
	post.Set("published", published)
	if published && !wasPublished {
		post.Set("published_at", time.Now())
	}

	if err := e.App.Save(post); err != nil {
		return e.JSON(400, map[string]string{"error": "Failed to update post"})
	}

	e.Response.Header().Set("HX-Redirect", "/admin/posts")
	return e.NoContent(200)
}

func handleDeletePost(e *core.RequestEvent) error {
	if _, err := isAdmin(e); err != nil {
		return e.Redirect(http.StatusSeeOther, "/auth/login")
	}

	postId := e.Request.PathValue("id")
	post, err := e.App.FindRecordById("posts", postId)
	if err != nil {
		return e.JSON(404, map[string]string{"error": "Post not found"})
	}

	if err := e.App.Delete(post); err != nil {
		return e.JSON(400, map[string]string{"error": err.Error()})
	}

	e.Response.Header().Set("HX-Redirect", "/admin/posts")
	return e.NoContent(200)
}

func handleAdminPostList(e *core.RequestEvent) error {
	log.Printf("handleAdminPostList called")
	if _, err := isAdmin(e); err != nil {
		log.Printf("isAdmin check failed: %v", err)
		return e.Redirect(http.StatusSeeOther, "/auth/login")
	}
	log.Printf("isAdmin check passed")

	posts, err := e.App.FindRecordsByFilter(
		"posts",
		"",
		"-published_at",
		50,
		0,
	)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		return e.JSON(500, map[string]string{"error": err.Error()})
	}

	adminPosts := make([]templates.AdminPost, 0, len(posts))
	for _, post := range posts {
		publishedAt := ""
		if !post.GetDateTime("published_at").IsZero() {
			publishedAt = post.GetDateTime("published_at").Time().Format("January 2, 2006")
		}
		adminPosts = append(adminPosts, templates.AdminPost{
			Id:          post.Id,
			Slug:        post.GetString("slug"),
			Title:       post.GetString("title"),
			Excerpt:     post.GetString("excerpt"),
			PublishedAt: publishedAt,
			Published:   !post.GetDateTime("published_at").IsZero(),
		})
	}

	csrfToken, err := middleware.GetCSRFToken(e)
	if err != nil {
		csrfToken = ""
	}

	return templates.AdminPostListPage(adminPosts, csrfToken).Render(e.Request.Context(), e.Response)
}
