package plugin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildPromptEmptyHistory(t *testing.T) {
	//Given
	var histories []History
	command := "pwd"

	//When
	prompt := buildPrompt(histories, command)

	//Then
	assert.Equal(t,
		"I want you to act as a Linux terminal. I will type commands and you will reply with what the terminal should show. I want you to only reply with the terminal output inside one unique code block, and nothing else. Do no write explanations. Do not type commands unless I instruct you to do so.\n\nA:pwd\n\nQ:/home/user\n\nA:pwd\n\nQ:",
		prompt)
}

func TestBuildPromptWithHistory(t *testing.T) {
	//Given
	var histories = []History{
		{
			Input:  "cat hello.txt",
			Output: "world",
		},
		{
			Input:  "echo 1234",
			Output: "1234",
		},
	}

	command := "pwd"

	//When
	prompt := buildPrompt(histories, command)

	//Then
	assert.Equal(t,
		"I want you to act as a Linux terminal. I will type commands and you will reply with what the terminal should show. I want you to only reply with the terminal output inside one unique code block, and nothing else. Do no write explanations. Do not type commands unless I instruct you to do so.\n\nA:pwd\n\nQ:/home/user\n\nA:cat hello.txt\n\nQ:world\n\nA:echo 1234\n\nQ:1234\n\nA:pwd\n\nQ:",
		prompt)
}
