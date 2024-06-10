## AI Chat: An Encore Application Connecting Chat Platforms with LLMs

This example application bridges the gap between popular chat platforms (Discord, Slack, local) and large language models (LLMs) like OpenAI and Google Gemini. It empowers you to create unique AI bots with distinct personas, capable of engaging in conversations with users and even amongst themselves.

**Features:**

- **Multi-platform Support:** Deploy bots on Discord, Slack, or a local web-based chat interface.
- **Personality Driven:** Craft engaging bots with unique personas and conversational styles.
- **LLM Flexibility:** Integrate with various LLMs, including OpenAI and Google Gemini.
- **Streamlined Development:** Built on Encore ([https://encore.dev/](https://encore.dev/)), simplifying development and deployment.

## Getting Started

### Prerequisites

- **Encore:** Install Encore by following the instructions at [https://encore.dev/docs/install](https://encore.dev/docs/install).
- **OpenAI API Key (Optional):** Obtain an API key from [https://platform.openai.com/api-keys](https://platform.openai.com/api-keys) if you want to utilize OpenAI's models.

### Quick Start

1. **Create Your Encore App:**
```bash
encore app create my-ai-chat --example=https://github.com/encoredev/ai-chat
```
(Replace `my-ai-chat` with your preferred app name)

2. **Set Your OpenAI API Key (Optional):**
```bash
cd my-ai-chat
encore secret set OpenAIKey --type dev,local,pr
```
Paste your OpenAI API key when prompted.

3. **Run Your App:**
```bash
encore run
```
Encore will build and start your application, providing you with the local URL (e.g., `http://127.0.0.1:4000`). Open this URL in your browser to interact with the local chat interface.

![Local chat interface demonstration](docs/assets/bots.gif)

### Deploy to the Cloud

Effortlessly deploy your application to a free testing environment on Encore's platform:

```bash
git push
```

Encore will guide you through the deployment process and provide a link to your live application.

![Encore deployment overview](docs/assets/deploy.png)

## System Architecture

The AI Chat application is designed as a modular system:

![System design diagram](docs/assets/system-design.png)

**Key Components:**

- **Chat Service:** The central coordinator, routing messages between chat platforms and the LLM service.
- **Discord Service:** Handles integration with the Discord API.
- **Slack Service:** Manages communication with the Slack API.
- **Local Service:** Provides a web-based chat interface for local testing and development.

- **Bot Service:** Responsible for creating, storing, and managing bot profiles (including personas).

- **LLM Service:** Formats prompts for LLMs, processes responses, and supports multiple LLM providers.
- **OpenAI Service:** Interfaces with OpenAI's API for chat completions and image generation.
- **Gemini Service:** Integrates with Google Gemini for chat completions.

## Message Flow

1. A user sends a message in a connected chat channel.
2. The corresponding chat integration (Discord, Slack, or local) receives the message.
3. The chat service identifies any bots present in the channel and retrieves their profiles along with the channel's message history.
4. The LLM service constructs a prompt tailored to the bot's persona and the conversation context.
5. The prompt is sent to the selected LLM provider (OpenAI or Gemini).
6. The LLM provider streams responses back to the LLM service.
7. The LLM service parses the responses and relays them to the chat service.
8. The chat service forwards the bot's responses to the appropriate chat channel.

## Chat Platform Integrations

**Local Chat Service:** A simple web interface for testing and local development, using WebSockets for real-time communication.

**Slack Service:** Utilizes the `slack-go` library ([github.com/slack-go/slack](github.com/slack-go/slack)) to interact with the Slack API. Note: You might need a service like ngrok ([https://ngrok.com/](https://ngrok.com/)) to expose your local development environment for testing.

**Discord Chat Service:** Employs the `discord-go` library ([github.com/bwmarrin/discordgo](github.com/bwmarrin/discordgo)) to connect to the Discord API and manage real-time communication through WebSockets.

## LLM Integrations

**OpenAI Service:** Leverages the `openai` library ([github.com/sashabaranov/go-openai](github.com/sashabaranov/go-openai)) to access OpenAI's API for chat completions and image generation (for bot avatars).

**Gemini Service:** Connects to Google Gemini using the `vertexai` library (`cloud.google.com/go/vertexai/genai`) for generating chat completions.

## Optional secrets

* `OpenAIKey`: The API key for OpenAI's services.
* `GeminiJSONCredentials`: The JSON credentials for a GCP service account with Google Gemini access.
* `SlackToken`: The OAuth token for your Slack app.
* `DiscordToken`: The token for your Discord bot.
* `NGrokToken`: The authentication token for ngrok (for local development).
* `NGrokDomain`: A domain for your ngrok tunnel (e.g., `my-ai-chat.ngrok.io`).

# Adding a slack bot
1. Create a Ngrok account (only required for local development)
Navigate to [https://ngrok.com/](https://ngrok.com/) and create an account.
Login to your account and add a tunnel auth token by clicking on `Authtokens` and then on `Add Tunnel Authtoken`.
![Create a auth token](docs/assets/ngrok-token.png)
Copy the token and add it as an encore secret
```bash
> encore secret set NGrokToken --type local
```
Then, create a ngrok domain by clicking Domains -> New Domain.
![Create a new ngrok domain](docs/assets/ngrok-domain.png)
Copy the domain name and add it as an encore secret
```bash
> encore secret set NGrokDomain --type local
```

2. Create a slack app
Navigate to [https://api.slack.com/apps](https://api.slack.com/apps) and click on "Create New App".
![Create a new slack app](docs/assets/slack-app.png)
Select `From an app manifest` and click on `Next`.
Next, pick the workspace where you want to develop your app in and click `Next`
Copy the [bot manifest](chat/provider/slack/bot-manifest.json) and paste it in the text box.
Replace the `<your-ngrok-domain>` with the domain you created in the previous step.
Click on `Next` and then on `Create`

3. Activate bot events
On the bot settings page, click "Event Subscription"
Start the encore app (this is necessary to activate event subscriptions):
```bash
> encore run
```
If the `Request URL` is yellow, click on `Retry`
![Enable event subscriptions](docs/assets/slack-events.png)

3. Install the app to your workspace
On the settings page, click `OAuth & Permissions` and then on `Install to Workspace`
[Authorize the app](docs/assets/slack-oauth.png)
On the OAuth permission page, select a channel where you want to add the bot and click `Allow`.
![Authorize the app](docs/assets/slack-channel.png)

4. Add the slack bot token
On the settings page, click on `OAuth & Permissions` and copy the `Bot User OAuth Token`.
Add the token as an encore secret
```bash
> encore secret set SlackToken --type local
```

5. Start the Encore App
Start the encore app
```bash
> encore run
```

6. Create a bot
In the [Encore Local dev dash](http://localhost:9400/), click on `API Explorer` 
and select the `bot.Create` endpoint. Add a name, prompt and enter `openai` as the llm:
![Create a bot](docs/assets/bot-create.gif)

Copy the bot id and save it for later

6. Find a slack channel id
In the [dev dash](http://localhost:9400/), click on `API Explorer` and select the `chat.ListChannels` endpoint.
and click `Call API`. In the returned list,  find the channel you want to add the bot to and copy the `id` field.

7. Add the bot to the slack channel
Select the `chat.AddBotToChannel` endpoint and input the bot id and the channel id you copied in the previous steps.
Click `Call API`.

8. Verify that the bot appears in your slack channel 
![slack-message.gif](docs/assets/slack-message.gif)

# Adding a discord bot
1. Create a discord bot
Go to [Developer Portal Applications](https://discord.com/developers/applications) and click on `New Application`.
Enter a name for you discord app and click on `Create`.
![discord-create-bot.png](docs/assets/discord-create-bot.png)
2. Configure Install Settings
Click on `Installation`. In `Install Link`, select `Discord Provided Link`. Then in
`Default Install settings`, add the `bot` scope and the following permissions:
* Connect
* Manage Web Hooks
* Read Message History
* Read Messages/View Channels
* Send Messages
![discord-install-settings.png](docs/assets/discord-install-settings.png)

3. Grant Privileged Gateway intents
Click on `Bot` and then on `Privileged Gateway Intents`. Enable the following intents:
* Server Members Intent
* Message Content Intent
![discord-intents.png](docs/assets/discord-intents.png)

4. Copy token
On the `Bot` page, click on `Reset Token` 

3. Install Bot
Copy the Install Link (e.g. `https://discord.com/oauth2/authorize?client_id=123123) and paste it in your browser.
Grant the bot access to a server by selecting a server and clicking `Continue' and then `Authorize`.
![discord-add-bot.png](docs/assets/discord-add-bot.png)
