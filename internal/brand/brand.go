// Package brand holds the product/developer information shown on the About
// page, footer and privacy policy (SRS §31, §33, §49).
package brand

const (
	Product        = "Fixed Deposit Management"
	Version        = "1.0.0"
	Developer      = "Aarti Tech Services"
	Services       = "FULL STACK DEVELOPMENT, SEO DIGITAL MARKETING, ENGINEERING SOLUTIONS"
	Website        = "https://aartitechservices.pages.dev/"
	Email          = "aartitechservices@gmail.com"
	Phone          = "+91 9869787575"
	Copyright      = "© 2026 · All rights reserved."
	Description    = "Simple and user-friendly software for managing Fixed Deposits, calculating interest and maturity amounts, handling renewals and closures, and generating Excel reports."
	DeveloperBlurb = "Build, grow, and transform your business with modern websites, custom software, SEO, digital marketing, and innovative technology solutions."
)

// Info is the payload returned to the About page.
type Info struct {
	Product        string `json:"product"`
	Version        string `json:"version"`
	Developer      string `json:"developer"`
	Services       string `json:"services"`
	Website        string `json:"website"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Copyright      string `json:"copyright"`
	Description    string `json:"description"`
	DeveloperBlurb string `json:"developerBlurb"`
}

// Current returns the static application information.
func Current() Info {
	return Info{
		Product:        Product,
		Version:        Version,
		Developer:      Developer,
		Services:       Services,
		Website:        Website,
		Email:          Email,
		Phone:          Phone,
		Copyright:      Copyright,
		Description:    Description,
		DeveloperBlurb: DeveloperBlurb,
	}
}
