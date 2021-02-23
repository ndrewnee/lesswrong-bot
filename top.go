package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func commandTop() (string, error) {
	archiveResponse, err := http.Get("https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10")
	if err != nil {
		return "", fmt.Errorf("get posts archive failed: %w", err)
	}

	var topPosts []Post

	if err := json.NewDecoder(archiveResponse.Body).Decode(&topPosts); err != nil {
		return "", fmt.Errorf("unmarshal top posts archive failed: %w", err)
	}

	text := bytes.NewBufferString("🏆 Top posts\n\n")

	for i, post := range topPosts {
		if post.Audience == "only_paid" {
			continue
		}

		text.WriteString(fmt.Sprintf("%v. [%s](%s)\n\n", i+1, post.Title, post.CanonicalURL))

		if post.Subtitle != "" && post.Subtitle != "..." {
			text.WriteString(fmt.Sprintf("    %s\n\n", post.Subtitle))
		}
	}

	return text.String(), nil
}
