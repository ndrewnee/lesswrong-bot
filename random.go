package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly"
)

const (
	DefaultLimit           = 12
	PostMaxLength          = 1500
	LesswrongPostsMaxCount = 25000
)

type (
	Post struct {
		Title string
		URL   string
		HTML  string
		Slug  string
	}

	AstralPost struct {
		Slug         string `json:"slug"`
		Title        string `json:"title"`
		Subtitle     string `json:"subtitle"`
		CanonicalURL string `json:"canonical_url"`
		BodyHTML     string `json:"body_html"`
		Audience     string `json:"audience"`
	}

	LesswrongResponse struct {
		Data LesswrongData `json:"data"`
	}

	LesswrongData struct {
		Posts LesswrongPost `json:"posts"`
	}

	LesswrongPost struct {
		Results []LesswrongResult `json:"results"`
	}

	LesswrongResult struct {
		Title    string        `json:"title"`
		PageURL  string        `json:"pageUrl"`
		HTMLBody string        `json:"htmlBody"`
		User     LesswrongUser `json:"user"`
	}

	LesswrongUser struct {
		DisplayName string `json:"displayName"`
	}
)

func (ap AstralPost) AsPost() Post {
	return Post{
		Title: ap.Title,
		URL:   ap.CanonicalURL,
		HTML:  ap.BodyHTML,
		Slug:  ap.Slug,
	}
}

func (lr LesswrongResult) AsPost() Post {
	return Post{
		Title: lr.Title,
		URL:   lr.PageURL,
		HTML:  lr.HTMLBody,
	}
}

func (b *Bot) CommandRandom(source Source) (string, error) {
	switch source {
	case SourceSlate:
		return b.CommandRandomSlate()
	case SourceAstral:
		return b.CommandRandomAstral()
	case SourceLesswrongRu:
		return b.CommandRandomLesswrongRu()
	case SourceLesswrong:
		return b.CommandRandomLesswrong()
	default:
		return b.CommandRandomSlate()
	}
}

func (b *Bot) CommandRandomSlate() (string, error) {
	// Load posts for the first time.
	if len(b.cache.slatePosts) == 0 {
		archivesCollector := colly.NewCollector()

		archivesCollector.OnHTML("a[href][rel=bookmark]", func(e *colly.HTMLElement) {
			b.cache.slatePosts = append(b.cache.slatePosts, Post{
				Title: e.Text,
				URL:   e.Attr("href"),
			})
		})

		if err := archivesCollector.Visit("https://slatestarcodex.com/archives/"); err != nil {
			return "", fmt.Errorf("get slatestarcodex posts failed: %s", err)
		}
	}

	if len(b.cache.slatePosts) == 0 {
		return "", fmt.Errorf("slatestarcodex posts not found")
	}

	i := b.randomInt(len(b.cache.slatePosts))
	post := b.cache.slatePosts[i]

	postCollector := colly.NewCollector()

	postCollector.OnHTML("div .entry-content", func(e *colly.HTMLElement) {
		post.HTML, _ = e.DOM.Html()
	})

	if err := postCollector.Visit(post.URL); err != nil {
		return "", fmt.Errorf("get slatestarcodex random post failed: %s", err)
	}

	return b.postToMarkdown(post)
}

func (b *Bot) CommandRandomAstral() (string, error) {
	// Load posts for the first time.
	if len(b.cache.astralPosts) == 0 {
		// As substack limits list to 12 posts in one request we fetch all posts using offset.
		for offset := 0; true; offset += DefaultLimit {
			uri := fmt.Sprintf("https://astralcodexten.substack.com/api/v1/archive?sort=new&limit=%d&offset=%d",
				DefaultLimit,
				offset,
			)

			httpResponse, err := b.httpClient.Get(uri)
			if err != nil {
				log.Println("[ERROR] Get astralcodexten posts failed: ", err)
				break
			}

			var newPosts []AstralPost

			if err := json.NewDecoder(httpResponse.Body).Decode(&newPosts); err != nil {
				log.Println("[ERROR] Unmarshal astralcodexten new posts failed: ", err)
				httpResponse.Body.Close()
				break
			}

			httpResponse.Body.Close()

			if len(newPosts) == 0 {
				break
			}

			for _, astralPost := range newPosts {
				if astralPost.Audience != "only_paid" {
					b.cache.astralPosts = append(b.cache.astralPosts, astralPost.AsPost())
				}
			}
		}
	}

	if len(b.cache.astralPosts) == 0 {
		return "", fmt.Errorf("astralcodexten posts not found")
	}

	i := b.randomInt(len(b.cache.astralPosts))
	post := b.cache.astralPosts[i]

	httpResponse, err := b.httpClient.Get("https://astralcodexten.substack.com/api/v1/posts/" + post.Slug)
	if err != nil {
		return "", fmt.Errorf("get astralcodexten random post failed: %s", err)
	}

	defer httpResponse.Body.Close()

	var astralPost AstralPost

	if err := json.NewDecoder(httpResponse.Body).Decode(&astralPost); err != nil {
		return "", fmt.Errorf("unmarshal astralcodexten post failed: %s", err)
	}

	return b.postToMarkdown(astralPost.AsPost())
}

func (b *Bot) CommandRandomLesswrongRu() (string, error) {
	// Load posts for the first time.
	if len(b.cache.lesswrongRuPosts) == 0 {
		postsCollector := colly.NewCollector()

		postsCollector.OnHTML("ul > li.leaf.menu-depth-4", func(e *colly.HTMLElement) {
			b.cache.lesswrongRuPosts = append(b.cache.lesswrongRuPosts, Post{
				Title: e.Text,
				URL:   e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
			})
		})

		if err := postsCollector.Visit("https://lesswrong.ru/w"); err != nil {
			return "", fmt.Errorf("get lesswrong.ru posts failed: %s", err)
		}
	}

	if len(b.cache.lesswrongRuPosts) == 0 {
		return "", fmt.Errorf("lesswrong.ru posts not found")
	}

	i := b.randomInt(len(b.cache.lesswrongRuPosts))
	post := b.cache.lesswrongRuPosts[i]

	postCollector := colly.NewCollector()

	postCollector.OnHTML("div.tex2jax", func(e *colly.HTMLElement) {
		post.HTML, _ = e.DOM.Html()
	})

	if err := postCollector.Visit(post.URL); err != nil {
		return "", fmt.Errorf("get lesswrong.ru random post failed: %s", err)
	}

	return b.postToMarkdown(post)
}

func (b *Bot) CommandRandomLesswrong() (string, error) {
	offset := b.randomInt(LesswrongPostsMaxCount)

	query := fmt.Sprintf(`{
		posts(input: {terms: {view: "new", limit: 1, offset: %d}}) {
			results {
				title
				pageUrl
				htmlBody
			}
		}
	}`, offset)

	request, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return "", fmt.Errorf("marshal request for lesswrong.com random post failed: %s", err)
	}

	httpResponse, err := b.httpClient.Post("https://www.lesswrong.com/graphql", "application/json", bytes.NewBuffer(request))
	if err != nil {
		return "", fmt.Errorf("get lesswrong.com random post failed: %s", err)
	}

	defer httpResponse.Body.Close()

	var response LesswrongResponse

	if err := json.NewDecoder(httpResponse.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("unmarshal lesswrong.com random post failed: %s", err)
	}

	if len(response.Data.Posts.Results) == 0 {
		return "", fmt.Errorf("lesswrong.com random post not found")
	}

	lesswrongPost := response.Data.Posts.Results[0]

	return b.postToMarkdown(lesswrongPost.AsPost())
}

func (b *Bot) postToMarkdown(post Post) (string, error) {
	markdown, err := b.mdConverter.ConvertString(post.HTML)
	if err != nil {
		return "", fmt.Errorf("convert lesswrong.ru html to markdown failed: %s", err)
	}

	// Cut post for preview mode.
	if len(markdown) > PostMaxLength {
		// Convert to runes to properly split between unicode symbols.
		r := []rune(markdown)

		// Truncate after next line end to not break markdown text.
		n := strings.IndexByte(string(r[PostMaxLength:]), '\n')
		if n != -1 {
			markdown = string(r[:PostMaxLength+n])
		} else {
			markdown = string(r[:PostMaxLength])
		}
	}

	// Stupid hotfixes for some invalid markdowns.
	markdown = strings.ReplaceAll(markdown, "* * *", "")
	markdown = strings.ReplaceAll(markdown, "```", "")
	markdown = strings.ReplaceAll(markdown, "![]", "[Image]")

	return fmt.Sprintf("📝 [%s](%s)\n\n%s", post.Title, post.URL, markdown), nil
}
