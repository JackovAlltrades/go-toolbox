package main

import (
	"fmt"
	"log"

	"github.com/JackovAlltrades/go-toolbox"
)

func main() {
	// Create a new instance of the toolbox
	var tools toolbox.Tools

	// Example strings to convert to slugs
	examples := []string{
		"This is a test",
		"Hello, World!",
		"Special Characters: æøå üöä ñ éèêë",
		"Multiple---hyphens and spaces",
		"Very long string that should be truncated because it exceeds the maximum length that we have set in our Slugify function which is 100 characters as defined in the implementation",
	}

	// Process each example
	for _, example := range examples {
		slug, err := tools.Slugify(example)
		if err != nil {
			log.Printf("Error slugifying '%s': %s", example, err)
			continue
		}
		
		fmt.Printf("Original: %s\nSlug: %s\n\n", example, slug)
	}

	// Test edge cases
	edgeCases := []string{
		"",                           // Empty string
		"!!!@@@###",                  // Only special characters
		"   ",                        // Only spaces
		"ÆØÅ",                        // Only non-ASCII characters
	}

	fmt.Println("Edge Cases:")
	for _, edge := range edgeCases {
		slug, err := tools.Slugify(edge)
		if err != nil {
			fmt.Printf("Original: '%s'\nError: %s\n\n", edge, err)
		} else {
			fmt.Printf("Original: '%s'\nSlug: %s\n\n", edge, slug)
		}
	}
}