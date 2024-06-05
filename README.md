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
encore secret set OpenAIKey --type dev,local,pr
```
When prompted for the secret, paste your [OpenAI API key](https://platform.openai.com/api-keys)  

Next, start the app
```bash
encore run
```

Navigate to the locally [hosted chat server](http://localhost:4000/), enter
your name and create a bot:

[insert video]

## Deploy to the cloud
Encore automatically triggers a deploy to a completely free testing environment if you push your app with git.

```bash
> git push
Enumerating objects: 29, done.
Counting objects: 100% (29/29), done.
Delta compression using up to 10 threads
Compressing objects: 100% (16/16), done.
Writing objects: 100% (16/16), 929.11 KiB | 7.62 MiB/s, done.
Total 16 (delta 11), reused 0 (delta 0), pack-reused 0
remote: main: triggered deploy https://app.encore.dev/my-app-name-regi/deploys/staging/deploy-id
To encore://slackbot-regi
   6d62c07..7d96015  main -> main
```

If you follow the deploy link in the command line output, you'll get to the encore deployment page
![docs/assets/deploy.png](docs/assets/deploy.png)

When the deploy has finished, click on the `Overview` link of the `staging` environment to see the URL of your deployed app.
![docs/assets/overview.png](docs/assets/overview.png)

To test your app
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
