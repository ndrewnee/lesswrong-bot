package models

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
		return "https://lesswrong.ru"
	case SourceSlate:
		return "https://slatestarcodex.com"
	case SourceAstral:
		return "https://astralcodexten.substack.com"
	case SourceLesswrong:
		return "https://lesswrong.com"
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
