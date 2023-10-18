package main

import (
	"context"
	"fmt"
	gogpt "github.com/sashabaranov/go-openai"
	"log"
)

func main() {
	config := gogpt.DefaultAzureConfig("sk-K2kmaIsLhynwlQIeCl4OXELsA4xyUe03eaRwFoIXR7w5Kqbz",
		"https://api.chatanywhere.com.cn")
	c := gogpt.NewClientWithConfig(config)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:     gogpt.GPT3Dot5Turbo,
		MaxTokens: 1000,
		Prompt:    "hello ",
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(resp.Choices[0].Text)
}
