package dictionary

// Definition represents a word definition from the dictionary API
type Definition struct {
	ID           string
	Text         string
	PartOfSpeech string
	Examples     []Example
}

// Example represents a usage example for a word
type Example struct {
	ID   string
	Text string
}

// DictionaryService defines the interface for dictionary services
type DictionaryService interface {
	GetDefinitions(word string) ([]Definition, error)
}
