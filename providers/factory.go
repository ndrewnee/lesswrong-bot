package providers

import (
	md "github.com/JohannesKaufmann/html-to-markdown"

	"github.com/ndrewnee/lesswrong-bot/interfaces"
	"github.com/ndrewnee/lesswrong-bot/models"
)

type ProviderFactory struct {
	storage     Storage
	httpClient  HTTPClient
	cacheExpire int
	randomInt   func(int) int
}

func NewProviderFactory(
	storage interfaces.Storage,
	httpClient interfaces.HTTPClient,
	cacheExpire int,
	randomInt func(int) int,
) *ProviderFactory {
	return &ProviderFactory{
		storage:     NewStorageAdapter(storage),
		httpClient:  NewHTTPClientAdapter(httpClient),
		cacheExpire: cacheExpire,
		randomInt:   randomInt,
	}
}

func (f *ProviderFactory) CreateProvider(source models.Source) PostProvider {
	switch source {
	case models.SourceLesswrongRu:
		return NewLessWrongRuProvider(f.storage, f.cacheExpire, f.randomInt)
	case models.SourceSlate:
		return NewSlateProvider(f.storage, f.cacheExpire, f.randomInt)
	case models.SourceAstral:
		return NewAstralProvider(f.storage, f.httpClient, f.cacheExpire, f.randomInt)
	case models.SourceLesswrong:
		return NewLessWrongProvider(f.httpClient, f.randomInt)
	default:
		return NewLessWrongRuProvider(f.storage, f.cacheExpire, f.randomInt)
	}
}

func (f *ProviderFactory) GetMarkdownConverter(source models.Source) *md.Converter {
	switch source {
	case models.SourceLesswrongRu:
		return md.NewConverter(models.DomainLesswrongRu, true, nil)
	case models.SourceSlate:
		return md.NewConverter(models.DomainSlate, true, nil)
	case models.SourceAstral:
		return md.NewConverter(models.DomainAstral, true, nil)
	case models.SourceLesswrong:
		return md.NewConverter(models.DomainLesswrong, true, nil)
	default:
		return md.NewConverter(models.DomainLesswrongRu, true, nil)
	}
}

func (f *ProviderFactory) ShouldUseURLWithText(source models.Source) bool {
	return source == models.SourceLesswrongRu
}
