package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// As https://slatestarcodex.com top posts won't change anymore it's much more effecient to return hardcoded list.
const MessageTopSlate = `🏆 Top posts from https://slatestarcodex.com

1. [Beware The Man Of One Study](https://slatestarcodex.com/2014/12/12/beware-the-man-of-one-study/)

2. [Meditations on Moloch](https://slatestarcodex.com/2014/07/30/meditations-on-moloch/)

3. [I Can Tolerate Anything Except The Outgroup](https://slatestarcodex.com/2014/09/30/i-can-tolerate-anything-except-the-outgroup/)

4. [Book Review: Albion’s Seed](https://slatestarcodex.com/2016/04/27/book-review-albions-seed/)

5. [Nobody Is Perfect, Everything Is Commensurable](https://slatestarcodex.com/2014/12/19/nobody-is-perfect-everything-is-commensurable/)

6. [The Control Group Is Out Of Control](https://slatestarcodex.com/2014/04/28/the-control-group-is-out-of-control/)

7. [Considerations On Cost Disease](https://slatestarcodex.com/2017/02/09/considerations-on-cost-disease/)

8. [Archipelago And Atomic Communitarianism](https://slatestarcodex.com/2014/06/07/archipelago-and-atomic-communitarianism/)

9. [The Categories Were Made For Man, Not Man For The Categories](https://slatestarcodex.com/2014/11/21/the-categories-were-made-for-man-not-man-for-the-categories/)

10. [Who By Very Slow Decay](https://slatestarcodex.com/2013/07/17/who-by-very-slow-decay/)`

func (b *Bot) CommandTop(source Source) (string, error) {
	switch source {
	case SourceSlate:
		return MessageTopSlate, nil
	case SourceAstral:
		return b.CommandTopAstral()
	default:
		return MessageTopSlate, nil
	}
}

func (b *Bot) CommandTopAstral() (string, error) {
	archiveResponse, err := b.httpClient.Get("https://astralcodexten.substack.com/api/v1/archive?sort=top&limit=10")
	if err != nil {
		return "", fmt.Errorf("get posts archive failed: %w", err)
	}

	var topPosts []AstralPost

	if err := json.NewDecoder(archiveResponse.Body).Decode(&topPosts); err != nil {
		return "", fmt.Errorf("unmarshal top posts archive failed: %w", err)
	}

	text := bytes.NewBufferString("🏆 Top posts from https://astralcodexten.substack.com\n\n")

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
