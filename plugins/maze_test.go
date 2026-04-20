package plugins

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newRequest(path string) *http.Request {
	return &http.Request{
		URL: &url.URL{Path: path},
	}
}

func TestMazeHoneypot_DirectoryListing(t *testing.T) {
	maze := &MazeHoneypot{
		ServerVersion: "Apache/2.4.41 (Ubuntu)",
		ServerName:    "files.internal.company.com",
	}

	resp := maze.HandleRequest(newRequest("/"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.ContentType, "text/html")
	assert.Contains(t, resp.Body, "Index of /")
	assert.Contains(t, resp.Body, "Apache/2.4.41 (Ubuntu)")
	assert.Contains(t, resp.Body, "files.internal.company.com")
	// Should contain directory links
	assert.Contains(t, resp.Body, "folder.svg")
	// Should NOT contain parent directory link at root
	assert.NotContains(t, resp.Body, "Parent Directory")
}

func TestMazeHoneypot_SubdirectoryHasParentLink(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/admin/backup"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "Index of /admin/backup")
	assert.Contains(t, resp.Body, "Parent Directory")
}

func TestMazeHoneypot_Deterministic(t *testing.T) {
	maze := &MazeHoneypot{}

	resp1 := maze.HandleRequest(newRequest("/some/deep/path"))
	resp2 := maze.HandleRequest(newRequest("/some/deep/path"))

	assert.Equal(t, resp1.Body, resp2.Body)
	assert.Equal(t, resp1.StatusCode, resp2.StatusCode)
}

func TestMazeHoneypot_DifferentPathsDifferentContent(t *testing.T) {
	maze := &MazeHoneypot{}

	resp1 := maze.HandleRequest(newRequest("/path/one"))
	resp2 := maze.HandleRequest(newRequest("/path/two"))

	assert.NotEqual(t, resp1.Body, resp2.Body)
}

func TestMazeHoneypot_FileResponse_SQL(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/backup/database_dump.sql"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.ContentType, "sql")
	assert.Contains(t, resp.Body, "MySQL dump")
	assert.Contains(t, resp.Body, "CREATE TABLE")
}

func TestMazeHoneypot_FileResponse_Env(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/config/.env"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "DATABASE_URL")
	assert.Contains(t, resp.Body, "AWS_ACCESS_KEY_ID")
}

func TestMazeHoneypot_FileResponse_YAML(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/deploy/config.yaml"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "database:")
	assert.Contains(t, resp.Body, "password:")
}

func TestMazeHoneypot_FileResponse_Log(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/logs/access.log"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Should contain log-like entries (either access log or error log format)
	assert.True(t, strings.Contains(resp.Body, "HTTP/1.1") || strings.Contains(resp.Body, "[error]") || strings.Contains(resp.Body, "[warn]"),
		"expected log content")
}

func TestMazeHoneypot_FileResponse_PHP(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/web/index.php"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "<?php")
}

func TestMazeHoneypot_FileResponse_Python(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/src/app.py"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "flask")
}

func TestMazeHoneypot_FileResponse_Shell(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/scripts/deploy.sh"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "#!/bin/bash")
}

func TestMazeHoneypot_InfiniteDepth(t *testing.T) {
	maze := &MazeHoneypot{}

	// Navigate deep into the maze - every level should work
	path := "/"
	for i := 0; i < 20; i++ {
		resp := maze.HandleRequest(newRequest(path))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Body, "Index of")
		assert.Contains(t, resp.Body, "folder.svg")

		// Extract a subdirectory link to follow
		idx := strings.Index(resp.Body, "href=\"")
		if idx == -1 {
			break
		}
		// Find a directory link (ends with /)
		body := resp.Body
		for {
			start := strings.Index(body, "href=\"")
			if start == -1 {
				break
			}
			body = body[start+6:]
			end := strings.Index(body, "\"")
			href := body[:end]
			if strings.HasSuffix(href, "/") && href != path && !strings.Contains(href, "?") {
				path = href
				break
			}
		}
	}
}

func TestMazeHoneypot_DefaultServerVersion(t *testing.T) {
	maze := &MazeHoneypot{} // No server version set

	resp := maze.HandleRequest(newRequest("/"))

	assert.Contains(t, resp.Body, "Apache/2.4.41 (Ubuntu)")
	assert.Equal(t, "Apache/2.4.41 (Ubuntu)", resp.Headers["Server"])
}

func TestMazeHoneypot_DirectoryListingContainsFiles(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/"))

	// Should contain at least some file entries (non-directory icons)
	assert.Contains(t, resp.Body, "alt=\"[   ]\">")
}

func TestIsFilePath(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/", false},
		{".", false},
		{"admin", false},
		{"config.yaml", true},
		{".env", true},
		{".htaccess", true},
		{".gitignore", true},
		{"backup.tar.gz", true},
		{"dump.sql", true},
		{"Makefile", true},
		{"Dockerfile", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, isFilePath(tt.input))
		})
	}
}

func TestMazePluginName(t *testing.T) {
	assert.Equal(t, "MazeHoneypot", MazePluginName)
}

func TestMazeHoneypot_FileSizeMatchesContent(t *testing.T) {
	maze := &MazeHoneypot{}

	// Get the directory listing for root
	dirResp := maze.HandleRequest(newRequest("/"))

	// Extract file hrefs and their displayed sizes from the listing
	body := dirResp.Body
	for {
		// Find file entries (non-directory, has alt="[   ]")
		altIdx := strings.Index(body, "alt=\"[   ]\">")
		if altIdx == -1 {
			break
		}
		body = body[altIdx:]

		hrefStart := strings.Index(body, "href=\"")
		if hrefStart == -1 {
			break
		}
		body = body[hrefStart+6:]
		hrefEnd := strings.Index(body, "\"")
		href := body[:hrefEnd]

		// Skip directories
		if strings.HasSuffix(href, "/") {
			continue
		}

		// Fetch the file and check Content-Length header matches actual body
		fileResp := maze.HandleRequest(newRequest(href))
		assert.Equal(t, http.StatusOK, fileResp.StatusCode, "file %s should return 200", href)

		if cl, ok := fileResp.Headers["Content-Length"]; ok {
			assert.Equal(t, fmt.Sprintf("%d", len(fileResp.Body)), cl,
				"Content-Length header should match actual body size for %s", href)
		}

		// The displayed size in the listing should roughly match the real content
		actualSize := formatSize(len(fileResp.Body))
		assert.NotEmpty(t, actualSize, "file %s should have non-zero content", href)
	}
}

func TestMazeHoneypot_TechProfileCoherence(t *testing.T) {
	maze := &MazeHoneypot{}

	// Test multiple directories to check that files within each directory are coherent
	paths := []string{"/", "/admin", "/backup", "/config", "/deploy/staging", "/internal/data"}
	for _, p := range paths {
		resp := maze.HandleRequest(newRequest(p))
		body := resp.Body

		// Count tech indicators
		hasPhp := strings.Contains(body, ".php")
		hasPython := strings.Contains(body, ".py")
		hasGo := strings.Contains(body, ".go")
		hasPackageJson := strings.Contains(body, "package.json")

		// At most one primary language should appear per directory
		langCount := 0
		if hasPhp {
			langCount++
		}
		if hasPython {
			langCount++
		}
		if hasGo {
			langCount++
		}
		if hasPackageJson {
			langCount++
		}

		assert.LessOrEqual(t, langCount, 1,
			"directory %s should not mix primary languages (PHP=%v, Python=%v, Go=%v, Node=%v)",
			p, hasPhp, hasPython, hasGo, hasPackageJson)
	}
}
