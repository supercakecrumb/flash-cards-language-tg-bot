package dictionary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// FreeDictionaryService implements the DictionaryService interface using the Free Dictionary API
type FreeDictionaryService struct {
	apiKey     string
	httpClient *http.Client
}

// FreeDictionaryResponse represents the response from the Free Dictionary API
type FreeDictionaryResponse []struct {
	Word      string `json:"word"`
	Phonetics []struct {
		Text  string `json:"text"`
		Audio string `json:"audio"`
	} `json:"phonetics"`
	Meanings []struct {
		PartOfSpeech string `json:"partOfSpeech"`
		Definitions  []struct {
			Definition string   `json:"definition"`
			Example    string   `json:"example"`
			Synonyms   []string `json:"synonyms"`
			Antonyms   []string `json:"antonyms"`
		} `json:"definitions"`
	} `json:"meanings"`
}

// NewFreeDictionaryService creates a new Free Dictionary service
func NewFreeDictionaryService(apiKey string) DictionaryService {
	return &FreeDictionaryService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDefinitions retrieves definitions for a word from the Free Dictionary API
func (s *FreeDictionaryService) GetDefinitions(word string) ([]Definition, error) {
	// Clean up the word
	word = strings.TrimSpace(strings.ToLower(word))

	// Build the API URL
	url := fmt.Sprintf("https://api.dictionaryapi.dev/api/v2/entries/en/%s", word)

	// Make the request
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, models.ErrExternalAPIError
	}

	// Parse the response
	var apiResp FreeDictionaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	// Convert to our Definition type
	var definitions []Definition

	for _, entry := range apiResp {
		for i, meaning := range entry.Meanings {
			for j, def := range meaning.Definitions {
				// Create a unique ID for the definition
				defID := fmt.Sprintf("%s_%d_%d", word, i, j)

				// Create definition
				definition := Definition{
					ID:           defID,
					Text:         def.Definition,
					PartOfSpeech: meaning.PartOfSpeech,
				}

				// Add example if available
				if def.Example != "" {
					exID := fmt.Sprintf("%s_%d_%d_ex", word, i, j)
					example := Example{
						ID:   exID,
						Text: def.Example,
					}
					definition.Examples = append(definition.Examples, example)
				}

				definitions = append(definitions, definition)
			}
		}
	}

	return definitions, nil
}
