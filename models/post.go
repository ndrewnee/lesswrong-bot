package models

const (
	DefaultLimit           = 12
	PostMaxLength          = 500
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
)

type (
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
