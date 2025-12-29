package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CezarGarrido/cachedpath"
)

func main() {
	fmt.Println("=== Usage examples of cachedpath library (Functional Options) ===")

	// Example 1: Simple download (no options)
	fmt.Println("1. Simple download via HTTPS:")
	path1, err := cachedpath.CachedPath(
		"https://raw.githubusercontent.com/golang/go/master/README.md",
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached file: %s\n\n", path1)
	}

	// Example 2: Local file
	fmt.Println("2. Local file verification:")
	path2, err := cachedpath.CachedPath("/etc/hosts")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Local file: %s\n\n", path2)
	}

	// Example 3: Download with custom cache directory
	fmt.Println("3. Download with custom cache directory:")
	path3, err := cachedpath.CachedPath(
		"https://golang.org/favicon.ico",
		cachedpath.WithCacheDir("/tmp/my_cache"),
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached file: %s\n\n", path3)
	}

	// Example 4: Download with custom timeout
	fmt.Println("4. Download with 60 second timeout:")
	path4, err := cachedpath.CachedPath(
		"https://httpbin.org/delay/2",
		cachedpath.WithTimeout(60*time.Second),
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached file: %s\n\n", path4)
	}

	// Example 5: Download with custom headers
	fmt.Println("5. Download with custom headers:")
	path5, err := cachedpath.CachedPath(
		"https://httpbin.org/user-agent",
		cachedpath.WithUserAgent("MyApp/2.0"),
		cachedpath.WithHeader("X-Custom-Header", "value"),
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached file: %s\n\n", path5)
	}

	// Example 6: Download with Bearer authentication
	fmt.Println("6. Download with Bearer authentication:")
	path6, err := cachedpath.CachedPath(
		"https://httpbin.org/bearer",
		cachedpath.WithAuth("my-secret-token"),
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached file: %s\n\n", path6)
	}

	// Example 7: Download with custom retry configuration
	fmt.Println("7. Download with 5 retries and 2 second delay:")
	path7, err := cachedpath.CachedPath(
		"https://golang.org/robots.txt",
		cachedpath.WithMaxRetries(5),
		cachedpath.WithRetryDelay(2*time.Second),
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached file: %s\n\n", path7)
	}

	// Example 8: Download in quiet mode
	fmt.Println("8. Download in quiet mode (no progress):")
	path8, err := cachedpath.CachedPath(
		"https://golang.org/LICENSE",
		cachedpath.WithQuiet(true),
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached file: %s\n\n", path8)
	}

	// Example 9: Download and archive extraction
	fmt.Println("9. Download and tar.gz file extraction:")
	path9, err := cachedpath.CachedPath(
		"https://github.com/golang/example/archive/refs/heads/master.tar.gz",
		cachedpath.WithExtractArchive(true),
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Extracted directory: %s\n\n", path9)
	}

	// Example 10: Access specific file inside archive
	fmt.Println("10. Access specific file inside archive:")
	path10, err := cachedpath.CachedPath(
		"https://github.com/golang/example/archive/refs/heads/master.tar.gz!example-master/README.md",
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Extracted file: %s\n\n", path10)
	}

	// Example 11: Custom HTTP client
	fmt.Println("11. Download with custom HTTP client:")
	customClient := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    50,
			IdleConnTimeout: 60 * time.Second,
		},
	}
	path11, err := cachedpath.CachedPath(
		"https://golang.org/doc/",
		cachedpath.WithHTTPClient(customClient),
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached file: %s\n\n", path11)
	}

	// Example 12: Combining multiple options
	fmt.Println("12. Combining multiple options:")
	path12, err := cachedpath.CachedPath(
		"https://httpbin.org/get",
		cachedpath.WithCacheDir("/tmp/advanced_cache"),
		cachedpath.WithTimeout(30*time.Second),
		cachedpath.WithMaxRetries(3),
		cachedpath.WithUserAgent("AdvancedApp/1.0"),
		cachedpath.WithHeader("Accept", "application/json"),
		cachedpath.WithQuiet(false),
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached file: %s\n\n", path12)
	}

	fmt.Println("=== Examples completed ===")
	fmt.Println("\nAdvantages of the Functional Options pattern:")
	fmt.Println("  • Cleaner and more readable code")
	fmt.Println("  • Easy to add new options without breaking compatibility")
	fmt.Println("  • Sensible default values")
	fmt.Println("  • Flexible composition of options")
}
