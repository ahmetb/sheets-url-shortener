## Redirector Server with Google Sheets

This is a simple web server that can redirect a pre-defined
set of URLs in Google Sheets, such as:

| shortcut | url |
|----|---|
| `gh` | `https://github.com/ahmetb/` |
| `book` | `https://docs.google.com/forms/d/e/1FAIpQLSefArw8NWiha6YCaoTccGZmo4QvuDYY4s87Y_tjW6h4al_4NQ/viewform` |

It can be deployed to [Google Cloud Run](https://cloud.run)
and run **free of charge** on its generous free tier (+Google
Sheets is free as well).

## Setup

1. Create a new **Google Sheet**: https://sheets.new.

1. Add two columns, first column is the "shortcut", the second
   column is the "url" to redirect the user. ([see example](#))

1. Save the ID of your Sheet from the URL (it’s a random string
   that looks like `1SMeoyesCaGHRlYdGj9VyqD-qhXtab1jrcgHZ0irvNDs`).

1. Click to deploy to Cloud Run, and provide your spreadsheet
   ID while deploying:

   [![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)

1. Go to https://console.cloud.google.com/run, click on
   `serverless-url-shortener` service. Find the email address written in the
   "Service Account" section.

1. Go to your Google Sheets, click "Share" and give this email
   address "Viewer" access on your sheet.

1. (Optional) If you want to use a custom domain like `go.ahmet.dev`, go to
   https://console.cloud.google.com/run/domains and map the
   `serverless-url-shortener` to your custom domain!
