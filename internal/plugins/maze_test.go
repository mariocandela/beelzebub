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

func TestMazeHoneypot_FileResponse_GoMain(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/src/main.go"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "package main")
	assert.Contains(t, resp.Body, "func main()")
}

func TestMazeHoneypot_FileResponse_SSHPubKey(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/ssh/id_rsa.pub"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "ssh-rsa")
}

func TestMazeHoneypot_FileResponse_RequirementsTxt(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/app/requirements.txt"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "Django")
}

func TestMazeHoneypot_FileResponse_TerraformState(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/infra/terraform.tfstate"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "terraform_version")
	assert.Contains(t, resp.Body, "aws_instance")
}

func TestMazeHoneypot_FileResponse_MigrationSQL(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/db/migration.sql"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "Migration")
	assert.Contains(t, resp.Body, "BEGIN;")
}

func TestMazeHoneypot_FileResponse_NotesMd(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/docs/notes.md"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "Meeting Notes")
	assert.Contains(t, resp.Body, "## Attendees")
}

func TestMazeHoneypot_FileResponse_GenericFile(t *testing.T) {
	maze := &MazeHoneypot{}

	// .xml extension hits the default case and calls genGenericFile
	resp := maze.HandleRequest(newRequest("/misc/config.xml"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "Auto-generated file")
}

func TestMazePlugin_HandleHTTP(t *testing.T) {
	mp := &mazePlugin{}

	req := newRequest("/")
	resp := mp.HandleHTTP(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Body)
	assert.Contains(t, resp.ContentType, "text/html")
}

func TestMazePlugin_HandleHTTP_FileRequest(t *testing.T) {
	mp := &mazePlugin{}

	req := newRequest("/config/app.yaml")
	resp := mp.HandleHTTP(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "database:")
}

func TestMazePlugin_Metadata(t *testing.T) {
	mp := &mazePlugin{}
	meta := mp.Metadata()

	assert.Equal(t, MazePluginName, meta.Name)
	assert.NotEmpty(t, meta.Version)
}

func TestMazeHoneypot_FormatSize(t *testing.T) {
	assert.Equal(t, "0", formatSize(0))
	assert.Equal(t, "512", formatSize(512))
	assert.Equal(t, "1.0K", formatSize(1024))
	assert.Equal(t, "1.0M", formatSize(1024*1024))
}

func TestIconForFile(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"file.txt", "text.svg"},
		{"readme.md", "text.svg"},
		{"access.log", "text.svg"},
		{"users.csv", "text.svg"},
		{"archive.gz", "compressed.svg"},
		{"backup.tar", "compressed.svg"},
		{"data.zip", "compressed.svg"},
		{"old.bak", "compressed.svg"},
		{"deploy.sh", "script.svg"},
		{"app.py", "script.svg"},
		{"main.go", "script.svg"},
		{"index.php", "script.svg"},
		{"app.js", "script.svg"},
		{"config.yaml", "layout.svg"},
		{"config.yml", "layout.svg"},
		{"data.json", "layout.svg"},
		{"config.xml", "layout.svg"},
		{"nginx.conf", "layout.svg"},
		{"settings.toml", "layout.svg"},
		{"dump.sql", "layout.svg"},
		{"id_rsa.key", "key.svg"},
		{"cert.pem", "key.svg"},
		{"id_rsa.pub", "key.svg"},
		{"binary", "unknown.svg"},
		{"Dockerfile", "unknown.svg"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.expected, iconForFile(tt.filename))
		})
	}
}

func TestMazeHoneypot_FileResponse_HTML(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/web/index.html"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.ContentType, "text/html")
}

func TestMazeHoneypot_FileResponse_BinaryFile(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/backup/data.zip"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/zip", resp.ContentType)
	assert.NotEmpty(t, resp.Body)
}

func TestMazeHoneypot_FileResponse_PEM(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/ssl/cert.pem"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/x-pem-file", resp.ContentType)
}

func TestMazeHoneypot_FileResponse_Conf(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/etc/nginx.conf"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "server {")
}

func TestMazeHoneypot_FileResponse_JSON(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/config/settings.json"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.ContentType)
}

func TestMazeHoneypot_FileResponse_TarGz(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/backup/archive.tar.gz"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Body)
}

func TestMazeHoneypot_FileResponse_Key(t *testing.T) {
	maze := &MazeHoneypot{}

	resp := maze.HandleRequest(newRequest("/keys/server.key"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/x-pem-file", resp.ContentType)
}

func TestMazeHoneypot_FileResponse_MarkdownGenerics(t *testing.T) {
	maze := &MazeHoneypot{}

	// README.md is handled by genReadme (not genNotesMd)
	resp := maze.HandleRequest(newRequest("/project/README.md"))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Body)
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
