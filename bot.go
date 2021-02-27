package main

import (
	"net/http"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

type (
	LesswrongBot struct {
		httpClient  HTTPClient
		mdConverter *md.Converter
		// Used as cache. Possible race. TODO Lock with mutex.
		astralPosts []AstralPost
		slatePosts  []SlatePost
	}

	HTTPClient interface {
		Get(uri string) (*http.Response, error)
	}
)

func NewLesswrongBot(httpClient HTTPClient, mdConverter *md.Converter) *LesswrongBot {
	return &LesswrongBot{
		httpClient:  httpClient,
		mdConverter: mdConverter,
	}
}
