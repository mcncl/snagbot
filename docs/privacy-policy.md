# SnagBot Privacy Policy

*Last Updated: March 7, 2025*

## Overview

SnagBot is a Slack bot that converts dollar amounts to equivalent items. This privacy policy describes how SnagBot collects, uses, and handles your information when you use our service.

## Information We Collect

SnagBot collects minimal information necessary to function:

1. **Channel Configuration**: We store your channel ID and custom configuration (item name and price) to provide our service.
2. **Workspace Information**: For multi-workspace installations, we store your Slack workspace ID, team name, and bot token.
3. **Message Content**: We process messages in channels where SnagBot is installed to detect dollar amounts, but we do not store message content.

## How We Use Information

We use the collected information solely to:
- Provide and maintain the SnagBot service
- Respond to dollar amounts with equivalent item calculations
- Apply channel-specific customizations

## Data Storage

- Channel configurations and workspace tokens are stored securely in Redis with appropriate encryption.
- We do not store message content after processing.
- Data is retained only for the duration necessary to provide the service.

## Data Sharing

We do not share your data with any third parties. We do not use your data for advertising or marketing purposes.

## Your Rights

You can reset your channel configuration at any time using the `/snagbot reset` command.

You can uninstall SnagBot at any time through your Slack workspace settings, which will remove all associated data.

## Changes to Privacy Policy

We may update this privacy policy from time to time. We will notify users of any significant changes by updating the "Last Updated" date at the top of this policy.

## Contact

If you have any questions about this privacy policy, please contact us at snags@benmcnicholl.com.
