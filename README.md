## URL Shortener with a Google Spreadsheet 📑

This is a simple web server that can redirect a pre-defined set of URLs in
Google Sheets. All you need to do is define a [Google Sheets][ex] like the
following, and deploy a simple app just by clicking around.

| shortcut | url |
|----|---|
| `gh` | `https://github.com/ahmetb` |
| `yt` | `https://www.youtube.com/watch?v=dQw4w9WgXcQ`
| `cv` | `https://docs.google.com/document/d/1c97PrqTFqiMpannv2/edit` |
| `cv` | `/resume` |

**Costs:** This can be deployed to [Google Cloud Run](https://cloud.run) and run
for **free** thanks to [generous free tier][free tier] (+Google Sheets is free
with a Gmail account as well).

[free tier]: https://cloud.google.com/run/pricing

This redirector also supports path additions on top of base URLs, for example,
with the sheet row:

| shortcut | url |
|----|---|
| `gcp` | `https://github.com/GoogleCloudPlatform` |

The `go.ahmet.dev/gcp/golang-samples` will be redirected to
`https://github.com/GoogleCloudPlatform/golang-samples`.

## Setup

1. Create a new **Google Sheet**: https://sheets.new.

1. Add two columns, first column is the "shortcut", the second
   column is the "url" to redirect the user. ([see example][ex])

1. Save the ID of your Sheet from the URL (it’s a random string
   that looks like `1SMeoyesCaGHRlYdGj9VyqD-qhXtab1jrcgHZ0irvNDs`).

1. Click to deploy to Cloud Run, and provide your spreadsheet
   ID while deploying:

   [![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)

1. Go to https://console.cloud.google.com/run, click on
   `sheets-url-shortener` service. Find the email address written in the
   "Service Account" section.

1. Go to your Google Sheets, click "Share" and give this email
   address "Viewer" access on your sheet.

1. Go to https://console.developers.google.com/apis/api/sheets.googleapis.com/overview
   and make sure the Google Sheets API is enabled.

1. (Optional) If you want to use a custom domain like `go.ahmet.dev`, go to
   https://console.cloud.google.com/run/domains and map the
   `sheets-url-shortener` to your custom domain!

## Advanced Configuration

This server can be configured with these following parameters:

| Environment Variable | Description |
|-----|------|
| `SHEET_NAME` | (optional) If you have multiple sheets in a spreadsheet, specify the sheet name. |
| `CACHE_TTL` | (optional) how frequently the spreadsheet must be reloaded (default: `5s`) |
| `HOME_REDIRECT` | (optional) which url to redirect when root url (`/`) is visited
| `LISTEN_ADDR` | (optional) which network address to listen on (default `""` which means all interfaces) |
| `PORT` | (optional) http port to listen on (default `8080`).

## Disclaimer

This is not an official Google project. It's distributed as-is under Apache 2.0
License. See [LICENSE](./LICENSE).

[ex]: https://docs.google.com/spreadsheets/d/1wCcj0Y4wFUHGAY0DmyefFlQ3s1DR8aUTz9seykM2iJ8/edit?usp=sharing
