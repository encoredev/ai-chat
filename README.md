# AI Chat Example App

This is an example app which connects chat services (discord, slack, local) with LLM models (openai, gemini, ...).
It supports creating bots with unique personas which will interact with themselves and the 
users in the channels they're added to.

## Prerequisites
* OpenAI token

## Quick Start

When you have [installed Encore](https://encore.dev/docs/install), you can create a new Encore application and clone this example with this command.

```bash
encore app create my-app-name --example=https://github.com/encoredev/ai-chat
```

Next, add your OpenAI token as a local secret:

```bash
cd my-app-name
encore secret set OpenAIKey -t local
```
When prompted for the secret, paste your [OpenAI API key](https://platform.openai.com/api-keys)  

Next, start the app
```bash
encore run
```

Navigate to the locally [hosted chat server](http://localhost:4000/encorechat/demo), enter
your name and create a bot:





## Running locally

Run your application:
```bash
encore run
```
To use the Slack integration, set the Slack signing secret (see tutorial above):
```bash
encore secret set SlackSigningSecret
```

## Open the developer dashboard

While `encore run` is running, open [http://localhost:9400/](http://localhost:9400/) to access Encore's [local developer dashboard](https://encore.dev/docs/observability/dev-dash).

## Deployment

Deploy your application to a staging environment in Encore's free development cloud:

```bash
git add -A .
git commit -m 'Commit message'
git push encore
```

Then head over to the [Cloud Dashboard](https://app.encore.dev) to monitor your deployment and find your production URL.

From there you can also connect your own AWS or GCP account to use for deployment.

Now off you go into the clouds!
