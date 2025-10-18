# Configuration Directory

This directory stores sensitive credentials and configuration for API integrations.

## Security Notice

**IMPORTANT**: Never commit actual credentials to git. All credential files are excluded via .gitignore.

## File Structure

### Quickbooks Configuration
- `quickbooks_credentials.json` - OAuth 2.0 credentials (gitignored)
- `quickbooks_tokens.json` - Access and refresh tokens (gitignored)

### Xero Configuration
- `xero_credentials.json` - OAuth 2.0 credentials (gitignored)
- `xero_tokens.json` - Access and refresh tokens (gitignored)

### Deputy Configuration
- `deputy_credentials.json` - API credentials (gitignored)
- `deputy_tokens.json` - Access tokens (gitignored)

## Template Files

Use the `.template` files as a starting point:

```bash
cp quickbooks_credentials.json.template quickbooks_credentials.json
# Then edit with your actual credentials
```

## Setting Up Credentials

### Quickbooks
1. Create an app at https://developer.intuit.com/
2. Get your Client ID and Client Secret
3. Set redirect URI (e.g., http://localhost:8000/callback)
4. Add credentials to `quickbooks_credentials.json`

### Xero
1. Create an app at https://developer.xero.com/
2. Get your Client ID and Client Secret
3. Set redirect URI
4. Add credentials to `xero_credentials.json`

### Deputy
1. Get API credentials from Deputy account settings
2. Add to `deputy_credentials.json`

## Environment Variables (Alternative)

You can also use environment variables instead of JSON files:

```bash
export QB_CLIENT_ID="your_client_id"
export QB_CLIENT_SECRET="your_client_secret"
export QB_REDIRECT_URI="http://localhost:8000/callback"

export XERO_CLIENT_ID="your_client_id"
export XERO_CLIENT_SECRET="your_client_secret"

export DEPUTY_API_KEY="your_api_key"
```
