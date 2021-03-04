package models

const (
	DomainLesswrongRu = "lesswrong.ru"
	DomainSlate       = "slatestarcodex.com"
	DomainAstral      = "astralcodexten.substack.com"
	DomainLesswrong   = "lesswrong.com"
)

const (
	SourceLesswrongRu Source = "1"
	SourceSlate       Source = "2"
	SourceAstral      Source = "3"
	SourceLesswrong   Source = "4"
)

type Source string

func (s Source) String() string {
	switch s {
	case SourceLesswrongRu:
		return "https://" + DomainLesswrongRu
	case SourceSlate:
		return "https://" + DomainSlate
	case SourceAstral:
		return "https://" + DomainAstral
	case SourceLesswrong:
		return "https://" + DomainLesswrong
	default:
		return ""
	}
}

func (s Source) Value() string {
	return string(s)
}

func (s Source) IsValid() bool {
	return s.String() != ""
}
