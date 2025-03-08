# Slack Bot Specification: "SnagBot"

## 1. Overview

The Slack bot (hereafter "SnagBot") is designed to automatically monitor messages in Slack channels where it is added. When a user mentions one or more dollar values (e.g., "We think this would cost $35"), the bot will reply in a thread with a fun conversion message. By default, the conversion is based on the cost of a Bunnings snag ($3.50 each). The bot sums any dollar amounts found in a message, calculates the number of snags by dividing the total by $3.50, rounds the result up, and responds with a message such as "That's nearly 10 Bunnings snags!"

Additionally, the bot will support configuration commands via Slack so that users can set a custom item and price (e.g., "coffee at $5.00") on a per-channel basis. If no custom item is set, Bunnings snags remain the default.

---

## 2. Functional Requirements

### 2.1 Message Monitoring & Response
- **Channel Scope:**  
  - The bot will only operate in channels where it has been explicitly added.
- **Dollar Value Detection:**  
  - The bot will scan each message for any mention of dollar values. This can be achieved using a regular expression (e.g., matching `\$[0-9]+(\.[0-9]{1,2})?`).
- **Multiple Dollar Values:**  
  - If a message includes multiple dollar values (e.g., "$35 for project A and $50 for project B"), the bot must total these amounts and use that sum for conversion.
- **Conversion Calculation:**  
  - Use the configured cost per item. By default, this is $3.50 (Bunnings snag).
  - Calculate the number of items by dividing the total dollar amount by the item's price.
  - **Rounding:** Always round up the result and include a qualifier such as "nearly" or "about" in the response.
- **Response Format:**  
  - The bot will reply in a threaded message.
  - The reply is a simple, fun message in the format:  
    "That's nearly X [item]!"  
    where X is the rounded-up number and [item] is either "Bunnings snags" (default) or the custom item configured for the channel.

### 2.2 Handling Message Edits
- **Message Update (Optional):**  
  - If a user edits a message that originally triggered the bot, the bot should attempt to recalculate and update its reply to reflect the new total.  
  - This feature is optional—if it proves too complex, it can be implemented later or only trigger for new messages.

### 2.3 Custom Configuration via Slack Commands
- **Custom Item Setup:**  
  - Users can configure a custom item and price using a Slack command.  
  - **Example command:**  
    `/snagbot set item "coffee" price 5.00`
- **Scope:**  
  - The configuration is per-channel. Different channels can have their own default item and price.
- **Validation:**  
  - The command should validate that:
    - The item name is provided as a string.
    - The price is provided in a valid numeric format (e.g., 5.00).
  - If validation fails, the bot should respond with an error message explaining the correct syntax.
- **Fallback:**  
  - If no custom configuration is set for a channel, the default "Bunnings snags" at $3.50 remains in use.

---

## 3. Architecture Choices

### 3.1 Slack API Integration
- **Event Subscriptions:**  
  - Use the Slack Events API to listen for new messages and message edits in channels where the bot is added.
- **Slash Commands:**  
  - Implement a slash command endpoint (e.g., `/snagbot`) to handle configuration commands.
- **Threaded Replies:**  
  - When responding, the bot posts its message as a thread reply to the original message.

### 3.2 Processing Flow
1. **Message Received:**  
   - The bot receives a message event.
2. **Detection:**  
   - It scans the message text for any occurrences of dollar values.
3. **Summation:**  
   - All detected values are summed.
4. **Conversion:**  
   - The sum is divided by the configured price (default $3.50 or the custom value) and rounded up.
5. **Response:**  
   - A reply is posted in the thread with the simple message format.

### 3.3 Customization Module
- **Command Parsing:**  
  - Parse and validate input from the configuration command.
- **Data Storage:**  
  - **In-Memory Storage:** Suitable for testing and small deployments.
  - **Database:** For production, use a persistent store (e.g., Redis, PostgreSQL) keyed by channel ID to store each channel's configuration.
- **Default Fallback:**  
  - If no configuration is found for a channel, automatically use "Bunnings snags" at $3.50.

---

## 4. Data Handling & Storage

- **Channel Configuration Data:**  
  - Each channel's configuration includes:
    - `item_name` (string, e.g., "Bunnings snags" or "coffee")
    - `price_per_item` (numeric, e.g., 3.50 or 5.00)
- **Storage Options:**
  - **In-Memory Cache:** For quick prototyping.
  - **Database:** Recommended for persistence across restarts. Use channel ID as the key.
- **Security Considerations:**  
  - Validate and sanitize all inputs.
  - Ensure that the bot only accesses channels it is a member of.

---

## 5. Error Handling Strategies

- **Dollar Value Parsing:**  
  - If no dollar values are found in a message, the bot should not respond.
- **Command Validation:**  
  - If a configuration command is malformed (e.g., missing item name or invalid price), reply with an error message outlining the correct syntax.
- **Calculation Errors:**  
  - Handle potential division errors gracefully.
- **Message Edit Edge Cases:**  
  - If the bot cannot update an existing thread (e.g., due to API issues), log the error and optionally notify an administrator.
- **API Rate Limits:**  
  - Monitor and respect Slack API rate limits, queueing messages if needed.

---

## 6. Testing Plan

### 6.1 Unit Testing
- **Dollar Value Extraction:**  
  - Test regular expression parsing for various formats (e.g., "$35", "$35.00", multiple values in one message).
- **Calculation Logic:**  
  - Verify that the conversion calculation correctly sums values and rounds up.
- **Configuration Command Parsing:**  
  - Test valid and invalid command inputs.
- **Fallback Behavior:**  
  - Confirm that channels with no custom configuration use the default values.

### 6.2 Integration Testing
- **Slack Event Simulation:**  
  - Simulate Slack messages and edits in a controlled test environment.
- **Slash Command Testing:**  
  - Simulate sending configuration commands and verify that the channel configuration is updated correctly.
- **Threaded Reply Verification:**  
  - Ensure the bot's replies appear in the correct message thread.

### 6.3 Error Case Testing
- **Malformed Input:**  
  - Confirm that the bot returns clear error messages when encountering invalid dollar amounts or configuration commands.
- **Edge Cases:**  
  - Test scenarios such as no dollar values, extremely large numbers, or messages with mixed text and numbers.

---

## 7. Developer Implementation Roadmap

1. **Slack App Setup:** ✅
   - Create a Slack App and set up event subscriptions for messages and message edits.
   - Set up slash command endpoints for configuration commands.

2. **Core Functionality:** ✅
   - Implement the message parser to detect dollar values.
   - Code the conversion logic (summing values, division, and rounding up).
   - Implement the threaded reply functionality.

3. **Configuration Module:** ✅
   - Develop the command parser for `/snagbot set item "..." price ...`.
   - Implement per-channel configuration storage (initially in-memory, then integrate with a database if required).

4. **Error Handling & Logging:** ✅
   - Integrate comprehensive error handling for message processing, API calls, and configuration parsing.
   - Set up logging for debugging and production monitoring.

5. **Testing & QA:** ✅
   - Write and run unit tests.
   - Conduct integration testing with Slack's API.
   - Gather user feedback in a staging channel before full deployment.

6. **Deployment & Monitoring:**
   - Deploy the bot to production.
   - Monitor performance and error logs.
   - Roll out additional features (e.g., automatic updates on message edits) as needed.