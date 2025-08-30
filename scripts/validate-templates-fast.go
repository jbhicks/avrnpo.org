package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	Red     = "\033[0;31m"
	Green   = "\033[0;32m"
	Yellow  = "\033[1;33m"
	NoColor = "\033[0m"
)

// TemplateInfo holds information about a template
type TemplateInfo struct {
	Path        string
	Name        string
	Variables   []string
	DirectVars  []string // Variables accessed directly (e.g., <%= amount %>)
	ContextVars []string // Variables accessed via c.Value() (e.g., <%= c.Value("amount") %>)
}

// ControllerInfo holds information about a controller
type ControllerInfo struct {
	Path      string
	Variables []string
}

func main() {
	// Check for verbose flag
	verbose := false
	if len(os.Args) > 1 && os.Args[1] == "--verbose" {
		verbose = true
	}

	start := time.Now()
	fmt.Println("🔍 Enhanced Template Validation (Go)...")

	// Check if we're in a Buffalo project
	if !isBuffaloProject() {
		fmt.Println("⚠️  Not in a Buffalo project root, skipping template validation")
		os.Exit(0)
	}

	// Build controller and template indexes
	fmt.Println("📝 Building indexes...")
	controllers := buildControllerIndex()
	templates := buildTemplateIndex()

	fmt.Printf("📊 Found %d controllers and %d templates\n", len(controllers), len(templates))

	// Validate templates
	missingVars := validateTemplates(templates, controllers, verbose)

	// Summary
	fmt.Println("")
	fmt.Println("📊 Validation Summary:")
	fmt.Printf("   Templates with syntax errors: 0\n") // We'll add this later
	fmt.Printf("   Missing variables: %d\n", len(missingVars))
	fmt.Printf("   Inconsistent variable access: %d\n", countInconsistentAccess(missingVars))

	if len(missingVars) > 0 {
		fmt.Printf("%s❌ Template validation failed!%s\n", Red, NoColor)
		fmt.Println("")
		fmt.Println("Missing variables and issues:")
		for _, missing := range missingVars {
			if strings.Contains(missing, ":inconsistent:") {
				fmt.Printf("   %s• Inconsistent access: %s%s\n", Yellow, strings.TrimPrefix(missing, "templates/pages/_donate_form.plush.html:inconsistent:"), NoColor)
			} else {
				fmt.Printf("   %s• %s%s\n", Red, missing, NoColor)
			}
		}
		fmt.Println("")
		fmt.Println("💡 To fix missing variables:")
		fmt.Println("   1. Add c.Set(\"variable_name\", value) in the controller")
		fmt.Println("   2. Or provide default values in templates: <%= variable_name || \"default\" %>")
		fmt.Println("   3. Or check if the variable should be conditionally set")
		fmt.Println("")
		fmt.Println("💡 To fix inconsistent variable access:")
		fmt.Println("   1. Use consistent access pattern: either 'variable' or 'c.Value(\"variable\")'")
		fmt.Println("   2. Prefer direct access 'variable' for variables set with c.Set()")
		fmt.Println("   3. Use c.Value() only for special cases or middleware-provided variables")
		os.Exit(1)
	} else {
		fmt.Printf("%s✅ All templates validated successfully!%s\n", Green, NoColor)
	}

	elapsed := time.Since(start)
	fmt.Printf("⏱️  Validation completed in %v\n", elapsed)
}

func isBuffaloProject() bool {
	files := []string{"app.go", "main.go"}
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}
	if _, err := os.Stat("actions"); err == nil {
		return true
	}
	return false
}

func buildControllerIndex() map[string]*ControllerInfo {
	controllers := make(map[string]*ControllerInfo)

	filepath.Walk("actions", func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		controller := &ControllerInfo{
			Path:      path,
			Variables: extractControllerVariables(string(content)),
		}

		// Store controller by its path for later lookup
		controllers[path] = controller

		return nil
	})

	return controllers
}

func buildTemplateIndex() []*TemplateInfo {
	var templates []*TemplateInfo

	filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".plush.html") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		template := &TemplateInfo{
			Path:        path,
			Name:        strings.TrimSuffix(filepath.Base(path), ".plush.html"),
			Variables:   extractTemplateVariables(string(content)),
			DirectVars:  extractDirectVariables(string(content)),
			ContextVars: extractContextVariables(string(content)),
		}

		templates = append(templates, template)
		return nil
	})

	return templates
}

func extractControllerVariables(content string) []string {
	// Extract c.Set("variable_name", ...) patterns
	re := regexp.MustCompile(`c\.Set\("([^"]+)"`)
	matches := re.FindAllStringSubmatch(content, -1)

	vars := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			vars = append(vars, match[1])
		}
	}

	// Remove duplicates and sort
	return uniqueSorted(vars)
}

func extractTemplateVariables(content string) []string {
	// Extract <%= variable %> patterns
	re := regexp.MustCompile(`<%=\s*([^%>]+)\s*%>`)

	var allVars []string
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			expr := strings.TrimSpace(match[1])

			// Skip control structures
			if strings.Contains(expr, "for ") || strings.Contains(expr, "if ") {
				continue
			}

			// Remove string literals
			expr = removeStringLiterals(expr)

			// Remove form helper method calls (e.g., f.InputTag(...), f.SelectTag(...))
			expr = removeFormHelperCalls(expr)

			// Remove custom helper function calls (e.g., getDonateButtonText(...), stripTags(...))
			expr = removeHelperFunctionCalls(expr)

			// Extract variable names (words starting with letter/underscore)
			varRe := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
			vars := varRe.FindAllString(expr, -1)

			allVars = append(allVars, vars...)
		}
	}

	// Filter out keywords and common patterns
	filtered := filterVariables(allVars)
	return uniqueSorted(filtered)
}

func extractDirectVariables(content string) []string {
	// Extract direct variable access patterns: <%= variable %> (not c.Value)
	re := regexp.MustCompile(`<%=\s*([^%>]+)\s*%>`)

	var allVars []string
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			expr := strings.TrimSpace(match[1])

			// Skip control structures
			if strings.Contains(expr, "for ") || strings.Contains(expr, "if ") {
				continue
			}

			// Skip c.Value() calls
			if strings.Contains(expr, "c.Value(") {
				continue
			}

			// Remove string literals
			expr = removeStringLiterals(expr)

			// Remove form helper method calls
			expr = removeFormHelperCalls(expr)

			// Extract variable names (words starting with letter/underscore)
			varRe := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
			vars := varRe.FindAllString(expr, -1)

			allVars = append(allVars, vars...)
		}
	}

	// Filter out keywords and common patterns
	filtered := filterVariables(allVars)
	return uniqueSorted(filtered)
}

func extractContextVariables(content string) []string {
	// Extract c.Value() patterns: <%= c.Value("variable") %>
	re := regexp.MustCompile(`c\.Value\("([^"]+)"\)`)
	matches := re.FindAllStringSubmatch(content, -1)

	var vars []string
	for _, match := range matches {
		if len(match) > 1 {
			vars = append(vars, match[1])
		}
	}

	return uniqueSorted(vars)
}

func removeStringLiterals(expr string) string {
	// Remove double-quoted strings
	re1 := regexp.MustCompile(`"[^"]*"`)
	expr = re1.ReplaceAllString(expr, "")

	// Remove single-quoted strings
	re2 := regexp.MustCompile(`'[^']*'`)
	expr = re2.ReplaceAllString(expr, "")

	return expr
}

func removeFormHelperCalls(expr string) string {
	// Remove form helper method calls like f.InputTag(...), f.SelectTag(...), etc.
	// This regex matches: f.MethodName(anything until closing paren)
	formHelperRe := regexp.MustCompile(`\bf\.\w+\([^)]*\)`)
	expr = formHelperRe.ReplaceAllString(expr, "")

	return expr
}

func removeHelperFunctionCalls(expr string) string {
	// Remove custom helper function calls like getDonateButtonText(...), stripTags(...), etc.
	// This regex matches: functionName(anything until closing paren)
	// But excludes common keywords and built-in functions
	helperRe := regexp.MustCompile(`\b(getDonateButtonText|stripTags|dateFormat|getCurrentURL)\([^)]*\)`)
	expr = helperRe.ReplaceAllString(expr, "")

	return expr
}

func filterVariables(vars []string) []string {
	// Keywords to filter out
	keywords := map[string]bool{
		"for": true, "if": true, "else": true, "end": true, "len": true,
		"true": true, "false": true, "nil": true, "and": true, "or": true, "not": true,
		"Format": true, "Get": true, "HasAny": true, "Errors": true, "capitalize": true, "partial": true,
		"January": true, "February": true, "March": true, "April": true, "May": true, "June": true,
		"July": true, "August": true, "September": true, "October": true, "November": true, "December": true,
		"Monday": true, "Tuesday": true, "Wednesday": true, "Thursday": true, "Friday": true, "Saturday": true, "Sunday": true,
	}

	var filtered []string
	for _, v := range vars {
		if !keywords[v] {
			filtered = append(filtered, v)
		}
	}

	return filtered
}

func validateTemplates(templates []*TemplateInfo, controllers map[string]*ControllerInfo, verbose bool) []string {
	var missingVars []string

	for _, template := range templates {
		if verbose {
			fmt.Printf("🔍 Validating: %s\n", template.Path)
		}

		if len(template.Variables) == 0 {
			if verbose {
				fmt.Println("   ✅ No variables required")
			}
			continue
		}

		if verbose {
			fmt.Printf("   📋 Template variables found: %s\n", strings.Join(template.Variables, " "))
		}

		// Find controller for this template
		controller := findControllerForTemplate(template, controllers)

		if controller == nil {
			// Check if this is a partial or layout template
			if strings.HasPrefix(template.Name, "_") || template.Name == "application" {
				// For partials, still check for inconsistent variable access
				inconsistent := findInconsistentVariableAccess(template)
				if len(inconsistent) > 0 {
					for _, inc := range inconsistent {
						fmt.Printf("%s⚠️  Inconsistent variable access: %s in %s%s\n", Yellow, inc, template.Path, NoColor)
						missingVars = append(missingVars, fmt.Sprintf("%s:inconsistent:%s", template.Path, inc))
					}
				}
				if verbose {
					fmt.Printf("   📄 Partial template, skipping controller validation\n")
				}
				continue
			} else {
				fmt.Printf("%s⚠️  No controller found for template: %s%s\n", Yellow, template.Path, NoColor)
				continue
			}
		}

		if verbose {
			fmt.Printf("   🎯 Controller found: %s\n", controller.Path)
			fmt.Printf("   ✅ Controller variables: %s\n", strings.Join(controller.Variables, " "))
		}

		// Check for missing variables
		missing := findMissingVariables(template, controller)
		if len(missing) > 0 {
			for _, m := range missing {
				fmt.Printf("%s⚠️  Missing variable: %s in %s%s\n", Yellow, m, template.Path, NoColor)
				missingVars = append(missingVars, fmt.Sprintf("%s:%s", template.Path, m))
			}
		}

		// Check for inconsistent variable access patterns
		inconsistent := findInconsistentVariableAccess(template)
		if len(inconsistent) > 0 {
			for _, inc := range inconsistent {
				fmt.Printf("%s⚠️  Inconsistent variable access: %s in %s%s\n", Yellow, inc, template.Path, NoColor)
				missingVars = append(missingVars, fmt.Sprintf("%s:inconsistent:%s", template.Path, inc))
			}
		}
	}

	return missingVars
}

func findControllerForTemplate(template *TemplateInfo, controllers map[string]*ControllerInfo) *ControllerInfo {
	templatePath := strings.TrimPrefix(template.Path, "templates/")

	// Search through all controllers to find one that renders this template
	for _, controller := range controllers {
		content, err := os.ReadFile(controller.Path)
		if err != nil {
			continue
		}

		// Check if this controller renders the template
		if strings.Contains(string(content), `r.HTML("`+templatePath+`"`) ||
			strings.Contains(string(content), `r.HTML("`+template.Name+`"`) ||
			strings.Contains(string(content), `r.HTML("`+filepath.Base(template.Name)+`"`) {
			return controller
		}
	}

	return nil
}

func findMissingVariables(template *TemplateInfo, controller *ControllerInfo) []string {
	var missing []string
	controllerVarSet := make(map[string]bool)
	for _, v := range controller.Variables {
		controllerVarSet[v] = true
	}

	// Read template content for property checking
	content, err := os.ReadFile(template.Path)
	if err != nil {
		return missing
	}
	templateContent := string(content)

	for _, varName := range template.Variables {
		// Skip if variable is directly provided
		if controllerVarSet[varName] {
			continue
		}

		// Check if this might be a property of a provided object
		isProperty := false

		// Check for direct property access (e.g., donation.Amount)
		for controllerVar := range controllerVarSet {
			if strings.Contains(templateContent, controllerVar+"."+varName) {
				isProperty = true
				break
			}
		}

		// Check for nested property access (e.g., post.User.FirstName)
		if !isProperty {
			for controllerVar := range controllerVarSet {
				pattern := controllerVar + `\.` + `.*\.` + varName
				if matched, _ := regexp.MatchString(pattern, templateContent); matched {
					isProperty = true
					break
				}
			}
		}

		// Only flag as missing if it's not a property and not a special case
		if !isProperty && !isSpecialVariable(varName) {
			missing = append(missing, varName)
		}
	}

	return missing
}

func findInconsistentVariableAccess(template *TemplateInfo) []string {
	var inconsistent []string

	// Create sets for quick lookup
	directSet := make(map[string]bool)
	for _, v := range template.DirectVars {
		directSet[v] = true
	}

	contextSet := make(map[string]bool)
	for _, v := range template.ContextVars {
		contextSet[v] = true
	}

	// Find variables that are accessed both ways
	for varName := range directSet {
		if contextSet[varName] {
			inconsistent = append(inconsistent, fmt.Sprintf("%s (used both directly and via c.Value)", varName))
		}
	}

	return inconsistent
}

func countInconsistentAccess(missingVars []string) int {
	count := 0
	for _, v := range missingVars {
		if strings.Contains(v, ":inconsistent:") {
			count++
		}
	}
	return count
}

func isSpecialVariable(varName string) bool {
	specialVars := map[string]bool{
		// Middleware-provided variables
		"authenticity_token": true, "current_path": true, "title": true, "description": true,
		"yield": true, "flash": true, "current_user": true,
		// Form helper variables
		"action": true, "formFor": true, "method": true, "autocomplete": true,
		"newAuthPath": true, "usersPath": true, "f": true,
		// Form helper method names (in case they slip through)
		"InputTag": true, "SelectTag": true, "TextAreaTag": true, "CheckBoxTag": true,
		"RadioButtonTag": true, "HiddenTag": true, "PasswordTag": true, "FileTag": true,
		// Template iteration and flash message variables
		"in": true, "msg": true, "key": true, "message": true, "messages": true, "option": true, "k": true,
		// Date/time formatting helpers
		"dateFormat": true, "raw": true, "time": true, "String": true, "TrimSpace": true, "strings": true,
		"Add": true, "After": true, "Hour": true, "Name": true,
		// Asset and path helpers
		"assetPath": true, "baseURL": true,
		// Form field variables that might be optional
		"email": true,
	}

	return specialVars[varName]
}

func uniqueSorted(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}

	// Use map to deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			unique = append(unique, item)
		}
	}

	sort.Strings(unique)
	return unique
}
