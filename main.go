package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	workdir := os.Getenv("WORKDIR")
	if workdir == "" {
		log.Fatal("WORKDIR environment variable is required")
	}

	// Set up logging to stderr (since stdout is used for MCP communication)
	log.SetOutput(os.Stderr)
	log.Printf("Starting go-mcp server...")

	s := server.NewMCPServer(
		"go-mcp",
		"0.1.0",
		server.WithToolCapabilities(true), // Enable tools
		server.WithLogging(),              // Add logging
	)
	godocServer := &GodocServer{
		Workdir: workdir,
		cache:   make(map[string]cachedDoc),
		server:  s,
	}

	/*
	   go doc --help
	   Usage of [go] doc:
	           go doc
	           go doc <pkg>
	           go doc <sym>[.<methodOrField>]
	           go doc [<pkg>.]<sym>[.<methodOrField>]
	           go doc [<pkg>.][<sym>.]<methodOrField>
	           go doc <pkg> <sym>[.<methodOrField>]
	   For more information run
	           go help doc

	   Flags:
	     -C dir
	           change to dir before running command
	     -all
	           show all documentation for package
	     -c    symbol matching honors case (paths not affected)
	     -cmd
	           show symbols with package docs even if package is a command
	     -short
	           one-line representation for each symbol
	     -src
	           show source code for symbol
	     -u    show unexported symbols as well as exported
	*/
	s.AddTool(mcp.Tool{
		Name:        "go_doc",
		Description: goDocToolDescription,
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"pkgSymMethodOrField": map[string]interface{}{
					"type":        "string",
					"description": " go doc <pkg> go doc <sym>[.<methodOrField>] go doc [<pkg>.]<sym>[.<methodOrField>] go doc [<pkg>.][<sym>.]<methodOrField> go doc <pkg> <sym>[.<methodOrField>]",
				},

				"cmd_flags": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "Optional: Additional go doc command flags. Common flags:\n" +
						"  -all: Show all documentation for package\n" +
						"  -src: Show the source code\n" +
						"  -u: Show unexported symbols as well as exported",
				},
			},
		},
	}, godocServer.handleGoDoc)

	/*
		go list --help
		usage: go list [-f format] [-json] [-m] [list flags] [build flags] [packages]
		Run 'go help list' for details.
	*/
	s.AddTool(mcp.Tool{
		Name:        "go_list",
		Description: "List packages or modules",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{

				"cmd_flags": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "Optional: Additional go list command flags. Common flags:\n" +
						"  -json: optional, print the output in JSON format. Usually not helpful for many packages due to large output.\n",
				},
				"packages": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "list of packages to list, github.com/user/repo, ./..., github.com/user/repo/..., github.com/user/repo/module/...",
				},
			},
		},
	}, godocServer.handleGoList)

	// Run server using stdio
	log.Printf("Starting stdio server...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	// Cleanup temporary directories before exit
	godocServer.cleanup()
}

const goDocToolDescription = `Get Go documentation for a package, type, function, or method.
This is the preferred and most efficient way to understand Go packages, providing official package
documentation in a concise format. Use this before attempting to read source files directly. Results
are cached and optimized for AI consumption.
The arguments are just like the 'go doc' command, with the package or symbol as the first argument.
`

type GodocServer struct {
	Workdir string
	server  *server.MCPServer
	cache   map[string]cachedDoc
}

type cachedDoc struct {
	content   string
	timestamp time.Time
	byteSize  int
}

// createTempProject creates a temporary Go project with the given package

// cleanup removes all temporary directories
func (s *GodocServer) cleanup() {

}

// runGoDoc executes the go doc command with the given arguments and optional working directory
func (s *GodocServer) runGoDoc(workingDir string, args ...string) (string, error) {
	// Create cache key that includes working directory
	cacheKey := workingDir + "|" + strings.Join(args, "|")

	// Check cache (with 5 minute expiration)
	// if doc, ok := s.cache[cacheKey]; ok {
	// 	if time.Since(doc.timestamp) < 5*time.Minute {
	// 		log.Printf("Cache hit for %s (%d bytes)", cacheKey, doc.byteSize)
	// 		return doc.content, nil
	// 	}
	// }

	cmd := exec.Command("go", append([]string{"doc"}, args...)...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Enhanced error handling with suggestions
		errStr := string(out)
		if strings.Contains(errStr, "no such package") || strings.Contains(errStr, "is not in std") {
			return "", fmt.Errorf("Package not found. Suggestions:\n"+
				"1. For standard library packages, use just the package name (e.g., 'io', 'net/http')\n"+
				"2. For external packages, ensure they are imported in the module\n"+
				"3. For local packages, provide the relative path (e.g., './pkg') or absolute path\n"+
				"4. Check for typos in the package name\n"+
				"Error details: %s", errStr)
		}
		if strings.Contains(errStr, "no such symbol") {
			return "", fmt.Errorf("Symbol not found. Suggestions:\n"+
				"1. Check if the symbol name is correct (case-sensitive)\n"+
				"2. Use -u flag to see unexported symbols\n"+
				"3. Use -all flag to see all package documentation\n"+
				"Error: %v", err)
		}
		if strings.Contains(errStr, "build constraints exclude all Go files") {
			return "", fmt.Errorf("No Go files found for current platform. Suggestions:\n"+
				"1. Try using -all flag to see all package files\n"+
				"2. Check if you need to set GOOS/GOARCH environment variables\n"+
				"Error: %v", err)
		}
		return "", fmt.Errorf("go doc error: %v\noutput: %s\nTip: Use -h flag to see all available options", err, errStr)
	}

	if len(out) == 0 {
		return "", fmt.Errorf("No documentation found by running the command: go doc %s from the directory: %s", strings.Join(args, " "), workingDir)
	}

	content := string(out)
	s.cache[cacheKey] = cachedDoc{
		content:   content,
		timestamp: time.Now(),
		byteSize:  len(content),
	}

	log.Printf("Cache miss for %s (%d bytes)", cacheKey, len(content))
	return content, nil
}

// handleGoDoc implements the tools/call endpoint
func (s *GodocServer) handleGoDoc(arguments map[string]interface{}) (result *mcp.CallToolResult, err error) {

	// Recover from any panics and return as error
	defer func() {
		if r := recover(); r != nil {

			result = &mcp.CallToolResult{
				IsError: true,
				Content: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": fmt.Sprintf("Error: %v", r),
					},
				},
			}
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	log.Printf("handleToolCall called with arguments: %+v", arguments)

	// Use the reso
	// Build command arguments
	var cmdArgs []string

	// Add any provided command flags
	if tmp, ok := getMapSliceAnyString(arguments, "cmd_flags"); ok && len(tmp) > 0 {
		cmdArgs = append(cmdArgs, tmp...)
	}

	if pkgSymMethodOrField, ok := getString(arguments, "pkgSymMethodOrField"); ok && pkgSymMethodOrField != "" {
		cmdArgs = append(cmdArgs, pkgSymMethodOrField)
	}

	// Run go doc command with working directory
	doc, err := s.runGoDoc(s.Workdir, cmdArgs...)
	if err != nil {
		return errResponse(err)
	}

	if doc == "" {
		doc = "No documentation found by go-mcp"
	}

	// Create the result with just the documentation
	result = &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": doc,
			},
		},
	}

	return result, nil
}

// handleGoList implements the tools/call endpoint
func (s *GodocServer) handleGoList(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	log.Printf("handleGoList called with arguments: %+v", arguments)

	// Add any provided command flags
	cmdArgs, ok := getMapSliceAnyString(arguments, "cmd_flags")
	if !ok {
		cmdArgs = []string{}
	}

	// Add package arguments
	packages, ok := getMapSliceAnyString(arguments, "packages")
	if ok {
		cmdArgs = append(cmdArgs, packages...)
	}

	// Run go list command with working directory
	cmd := exec.Command("go", append([]string{"list"}, cmdArgs...)...)
	if s.Workdir != "" {
		cmd.Dir = s.Workdir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go list error: %v\noutput: %s", err, string(out))
	}

	// Create the result with just the documentation
	result := &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": string(out),
			},
		},
	}

	return result, nil
}

func getMapSliceAny(m map[string]interface{}, key string) ([]interface{}, bool) {
	if v, ok := m[key]; ok {
		if a, ok := v.([]interface{}); ok {
			return a, true
		}
	}
	return nil, false
}

func getMapSliceAnyString(m map[string]interface{}, key string) ([]string, bool) {
	// try to get as []string
	res, ok := func() ([]string, bool) {
		if v, ok := m[key]; ok {
			if a, ok := v.([]string); ok {
				return a, true
			}
		}
		return nil, false
	}()
	if ok {
		return res, true
	}

	if v, ok := m[key]; ok {
		if a, ok := v.([]interface{}); ok {
			tmp := make([]string, len(a))
			for i, x := range a {
				if s, ok := x.(string); ok {
					tmp[i] = s
				} else {
					return nil, false
				}
			}
			return tmp, true
		}
	}
	return nil, false
}

func getString(m map[string]interface{}, key string) (string, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func errResponse(err error) (*mcp.CallToolResult, error) {
	if err == nil {
		return nil, nil
	}
	return &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": fmt.Sprintf("Error: %v", err),
			},
		},
		IsError: true,
	}, nil
}
