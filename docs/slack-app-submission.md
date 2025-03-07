# Slack App Directory Submission Guide for SnagBot

This document outlines the steps required to submit SnagBot to the Slack App Directory.

## Prerequisites

1. **Required Documents**:
   - Privacy Policy: `/docs/privacy-policy.md` - Must be hosted on a public URL
   - Terms of Service: `/docs/terms-of-service.md` - Must be hosted on a public URL
   - App Logo: Create a 512x512px PNG with a transparent background

2. **Hosting Requirements**:
   - HTTPS endpoint for your app
   - Redis instance configured (AWS ElastiCache setup in `/terraform`)
   - Properly configured DNS for your domain

3. **App Configuration**:
   - Update `manifest.json` with your actual domain names
   - Configure OAuth redirect URLs in your Slack App dashboard

## Configuration Steps

1. **Set Up Your Hosting Environment**:
   ```bash
   # Deploy Redis instance
   cd terraform
   ./deploy.sh
   
   # Note the Redis URL from the output
   ```

2. **Set Environment Variables**:
   ```bash
   export PORT=8080
   export SLACK_SIGNING_SECRET=<your-signing-secret>
   export SLACK_CLIENT_ID=<your-client-id>
   export SLACK_CLIENT_SECRET=<your-client-secret>
   export REDIS_URL=<your-redis-url-from-terraform>
   export APP_BASE_URL=<your-public-app-url>
   export COOKIE_SECRET=<random-secure-string>
   export JWT_SECRET=<random-secure-string>
   ```

3. **Update the Manifest**:
   - Edit `manifest.json` to replace `YOUR_DOMAIN_HERE` with your actual domain
   - In your Slack App Dashboard, under "App Manifest," replace the existing JSON with your updated manifest

## Submission Checklist

- [ ] App icon (512x512px PNG with transparent background)
- [ ] Privacy Policy hosted at a public URL
- [ ] Terms of Service hosted at a public URL
- [ ] Support email address set in Slack App dashboard
- [ ] App Description and detailed explanation of functionality
- [ ] At least one screenshot of the app in action
- [ ] Verification that all required scopes are properly configured
- [ ] Successful test installations on at least 2 different workspaces
- [ ] All environment variables properly set in production
- [ ] Redis configured and working in production
- [ ] OAuth flow tested and working

## App Review Process

Once you submit your app for review, the Slack team will:

1. Review your app's functionality
2. Verify your privacy policy and terms of service
3. Test the installation flow
4. Check for security best practices
5. Ensure compliance with Slack's guidelines

The review process typically takes 1-2 weeks. You may be contacted for additional information or changes before approval.

## After Approval

Once approved, you'll need to:

1. Set up analytics to track installations
2. Monitor app performance
3. Set up support channels for user questions
4. Consider a versioning strategy for future updates

## Resources

- [Slack API Documentation](https://api.slack.com/docs)
- [App Directory Requirements](https://api.slack.com/start/distributing/guidelines)
- [Slack App Security Guidelines](https://api.slack.com/authentication/best-practices)