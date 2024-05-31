package llm

import (
	_ "embed"
)

// These embeds are used to store prompts defined in the prompts/ directory.

//go:embed prompts/reply.txt
var replyPrompt []byte

//go:embed prompts/intro.txt
var introPrompt []byte

//go:embed prompts/goodbye.txt
var goodbyePrompt []byte

//go:embed prompts/avatar.txt
var avatarPrompt []byte

//go:embed prompts/persona.txt
var personaPrompt []byte

//go:embed prompts/response.txt
var responsePrompt []byte

//go:embed prompts/create_persona.txt
var botPrompt []byte
