// brandInfo.ts — static fallback copy of the branding until AppInfo() loads.
// The About page always prefers the value returned by the backend.

export const brand = {
  product: "Fixed Deposit Management",
  version: "1.0.0",
  developer: "Aarti Tech Services",
  services:
    "FULL STACK DEVELOPMENT, SEO DIGITAL MARKETING, ENGINEERING SOLUTIONS",
  website: "https://aartitechservices.pages.dev/",
  email: "aartitechservices@gmail.com",
  phone: "+91 9869787575",
  copyright: "© 2026 · All rights reserved.",
  description:
    "Simple and user-friendly software for managing Fixed Deposits, calculating interest and maturity amounts, handling renewals and closures, and generating Excel reports.",
  developerBlurb:
    "Build, grow, and transform your business with modern websites, custom software, SEO, digital marketing, and innovative technology solutions.",
} as const;
