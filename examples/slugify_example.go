package main

import (
    "fmt"
    "github\.com/JackovAlltrades/go-toolbox"
)

func main() {
    tools := &toolbox.Tools{}
    
    examples := []string{
        "Hello World!",
        "This is a test",
        "Special Characters: !@#$%^&*()",
        "Multiple   Spaces",
        "Dashes-and_underscores",
    }
    
    for _, example := range examples {
        slug, err := tools.Slugify(example)
        if err != nil {
            fmt.Printf("Error slugifying '%s': %v\n", example, err)
            continue
        }
        
        fmt.Printf("Original: '%s' -> Slug: '%s'\n", example, slug)
    }
}

