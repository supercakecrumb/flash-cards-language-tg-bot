package dictionary

// MockDictionaryService implements the DictionaryService interface for testing
type MockDictionaryService struct {
	Definitions map[string][]Definition
}

// NewMockDictionaryService creates a new mock dictionary service
func NewMockDictionaryService() *MockDictionaryService {
	return &MockDictionaryService{
		Definitions: make(map[string][]Definition),
	}
}

// GetDefinitions returns mock definitions for a word
func (s *MockDictionaryService) GetDefinitions(word string) ([]Definition, error) {
	if defs, ok := s.Definitions[word]; ok {
		return defs, nil
	}

	// Return some default mock data if word not found
	return []Definition{
		{
			ID:           "mock_1",
			Text:         "A mock definition for testing purposes",
			PartOfSpeech: "noun",
			Examples: []Example{
				{
					ID:   "mock_ex_1",
					Text: "This is a mock example",
				},
			},
		},
	}, nil
}

// AddMockDefinition adds a mock definition for a word
func (s *MockDictionaryService) AddMockDefinition(word string, definition Definition) {
	if _, ok := s.Definitions[word]; !ok {
		s.Definitions[word] = []Definition{}
	}
	s.Definitions[word] = append(s.Definitions[word], definition)
}
