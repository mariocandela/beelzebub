package plugins

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/http"
	"path"
	"strings"
	"time"
)

const MazePluginName = "MazeHoneypot"

// MazeHoneypot generates an infinite, deterministic graph of Apache-style
// directory listings. Every path resolves to either a directory (more links)
// or a file (realistic content). The same URL always produces the same page.
type MazeHoneypot struct {
	ServerVersion string // e.g. "Apache/2.4.41 (Ubuntu)"
	ServerName    string // hostname shown in footer
}

// MazeResponse holds the generated HTTP response for a maze request.
type MazeResponse struct {
	StatusCode  int
	ContentType string
	Body        string
	Headers     map[string]string
}

// dirNames are realistic directory names found on web servers.
var dirNames = []string{
	"admin", "backup", "backups", "config", "conf", "data", "database",
	"db", "deploy", "dev", "docs", "downloads", "dump", "export",
	"files", "home", "images", "import", "includes", "internal",
	"legacy", "lib", "log", "logs", "mail", "media", "migration",
	"old", "private", "public", "reports", "resources", "scripts",
	"secret", "server", "shared", "src", "staging", "static",
	"storage", "system", "temp", "test", "tools", "tmp", "upload",
	"uploads", "users", "var", "vendor", "web", "www", "archive",
	"assets", "bin", "build", "cache", "certs", "cgi-bin", "clients",
	"cms", "core", "credentials", "dashboard", "debug", "dist",
	"docker", "env", "etc", "frontend", "gateway", "git", "hooks",
	"infra", "jenkins", "keys", "kubernetes", "lambda", "monitoring",
	"node_modules", "ops", "output", "packages", "patches", "pipeline",
	"production", "provisioning", "recovery", "release", "repo",
	"setup", "snapshots", "sql", "ssl", "swagger", "terraform",
	"tokens", "vault", "workspace",
}

// fileTemplate defines a single file entry with its generator.
type fileTemplate struct {
	name    string
	ext     string
	genFunc func(r *rand.Rand, fullPath string) string
}

// techProfile groups files that would realistically appear together in the
// same project directory, preventing incoherent tech-stack mixing.
type techProfile struct {
	name  string
	files []fileTemplate
}

// techProfiles defines coherent sets of files per technology stack.
// When generating a directory listing, one profile is selected deterministically
// so that all files in a directory belong to the same realistic project.
var techProfiles = []techProfile{
	{
		name: "wordpress-php",
		files: []fileTemplate{
			{"wp-config", ".php", genWPConfig},
			{"index", ".php", genPHPIndex},
			{".htaccess", "", genHtaccess},
			{"database_dump", ".sql", genSQLDump},
			{"backup", ".tar.gz", genBinaryPlaceholder},
			{".env", "", genEnvFile},
			{"error", ".log", genErrorLog},
			{"access", ".log", genAccessLog},
			{"users", ".csv", genUsersCSV},
			{"nginx", ".conf", genNginxConf},
		},
	},
	{
		name: "python-flask",
		files: []fileTemplate{
			{"app", ".py", genPythonApp},
			{"requirements", ".txt", genRequirementsTxt},
			{"config", ".yaml", genYAMLConfig},
			{".env", "", genEnvFile},
			{"Dockerfile", "", genDockerfile},
			{"docker-compose", ".yml", genDockerCompose},
			{"migration", ".sql", genMigrationSQL},
			{"README", ".md", genReadme},
			{".gitignore", "", genGitignore},
			{"deploy", ".sh", genDeployScript},
			{"access", ".log", genAccessLog},
		},
	},
	{
		name: "go-service",
		files: []fileTemplate{
			{"main", ".go", genGoMain},
			{"Makefile", "", genMakefile},
			{"config", ".yaml", genYAMLConfig},
			{".env", "", genEnvFile},
			{"Dockerfile", "", genDockerfile},
			{"docker-compose", ".yml", genDockerCompose},
			{"README", ".md", genReadme},
			{".gitignore", "", genGitignore},
			{"deploy", ".sh", genDeployScript},
			{"terraform", ".tfstate", genTerraformState},
		},
	},
	{
		name: "node-express",
		files: []fileTemplate{
			{"package", ".json", genPackageJSON},
			{"settings", ".json", genJSONConfig},
			{".env", "", genEnvFile},
			{"Dockerfile", "", genDockerfile},
			{"docker-compose", ".yml", genDockerCompose},
			{"README", ".md", genReadme},
			{".gitignore", "", genGitignore},
			{"deploy", ".sh", genDeployScript},
			{"access", ".log", genAccessLog},
			{"nginx", ".conf", genNginxConf},
		},
	},
	{
		name: "infra-ops",
		files: []fileTemplate{
			{"terraform", ".tfstate", genTerraformState},
			{"docker-compose", ".yml", genDockerCompose},
			{"deploy", ".sh", genDeployScript},
			{"nginx", ".conf", genNginxConf},
			{"credentials", ".txt", genCredentialsTxt},
			{"id_rsa", ".pub", genSSHPubKey},
			{"server", ".key", genBinaryPlaceholder},
			{"notes", ".md", genNotesMd},
			{"todo", ".txt", genTodoTxt},
			{"config", ".yaml", genYAMLConfig},
			{"backup", ".tar.gz", genBinaryPlaceholder},
		},
	},
	{
		name: "database-backup",
		files: []fileTemplate{
			{"database_dump", ".sql", genSQLDump},
			{"migration", ".sql", genMigrationSQL},
			{"backup", ".tar.gz", genBinaryPlaceholder},
			{"users", ".csv", genUsersCSV},
			{"config", ".yaml", genYAMLConfig},
			{".env", "", genEnvFile},
			{"README", ".md", genReadme},
			{"restore", ".sh", genDeployScript},
			{"error", ".log", genErrorLog},
			{"credentials", ".txt", genCredentialsTxt},
		},
	},
}

// allFileTemplates is a flat list used only for file-response lookups by extension.
var allFileTemplates = func() []fileTemplate {
	seen := make(map[string]bool)
	var result []fileTemplate
	for _, profile := range techProfiles {
		for _, f := range profile.files {
			key := f.name + f.ext
			if !seen[key] {
				seen[key] = true
				result = append(result, f)
			}
		}
	}
	return result
}()

// seedFromPath returns a deterministic int64 seed derived from the given path.
func seedFromPath(p string) int64 {
	h := sha256.Sum256([]byte(p))
	return int64(binary.BigEndian.Uint64(h[:8]))
}

// HandleRequest generates a maze response for the given HTTP request path.
func (m *MazeHoneypot) HandleRequest(request *http.Request) MazeResponse {
	reqPath := path.Clean(request.URL.Path)
	if reqPath == "." {
		reqPath = "/"
	}

	// Determine if this looks like a file request (has an extension or is a known dotfile)
	base := path.Base(reqPath)
	if isFilePath(base) {
		return m.generateFileResponse(reqPath)
	}
	return m.generateDirectoryListing(reqPath)
}

// knownNoExtFiles are files with no extension that should be served as files, not directories.
var knownNoExtFiles = map[string]bool{
	"Dockerfile": true,
	"Makefile":   true,
}

// isFilePath returns true if the basename looks like a file (has extension or is a dotfile).
func isFilePath(base string) bool {
	if base == "/" || base == "." {
		return false
	}
	if knownNoExtFiles[base] {
		return true
	}
	// dotfiles like .env, .htaccess, .gitignore
	if strings.HasPrefix(base, ".") && !strings.Contains(base[1:], "/") {
		ext := path.Ext(base)
		if ext == "" {
			return true
		}
	}
	ext := path.Ext(base)
	return ext != ""
}

func (m *MazeHoneypot) generateDirectoryListing(reqPath string) MazeResponse {
	r := rand.New(rand.NewSource(seedFromPath(reqPath)))

	// Generate subdirectories (3-7)
	numDirs := 3 + r.Intn(5)
	// Generate files (2-6)
	numFiles := 2 + r.Intn(5)

	// Pick unique directory names
	shuffled := make([]string, len(dirNames))
	copy(shuffled, dirNames)
	r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	// Add depth-specific prefixes to some dirs to make paths more varied
	depth := strings.Count(reqPath, "/")
	dirs := make([]string, 0, numDirs)
	for i := 0; i < numDirs && i < len(shuffled); i++ {
		name := shuffled[i]
		// Occasionally add a version suffix or date suffix
		roll := r.Intn(10)
		if roll < 2 && depth < 8 {
			name = fmt.Sprintf("%s_%d", name, 2020+r.Intn(7))
		} else if roll < 4 && depth < 8 {
			name = fmt.Sprintf("%s_v%d", name, 1+r.Intn(5))
		}
		dirs = append(dirs, name)
	}

	// Pick a coherent tech profile for this directory so files make sense together
	profile := techProfiles[r.Intn(len(techProfiles))]
	profileFiles := make([]fileTemplate, len(profile.files))
	copy(profileFiles, profile.files)
	r.Shuffle(len(profileFiles), func(i, j int) { profileFiles[i], profileFiles[j] = profileFiles[j], profileFiles[i] })

	type fileEntry struct {
		name string
		size string
	}
	files := make([]fileEntry, 0, numFiles)
	for i := 0; i < numFiles && i < len(profileFiles); i++ {
		tmpl := profileFiles[i]
		fname := tmpl.name + tmpl.ext
		// Compute real content size so the listing matches the actual Content-Length
		filePath := path.Join(reqPath, fname)
		fileR := rand.New(rand.NewSource(seedFromPath(filePath)))
		body := tmpl.genFunc(fileR, filePath)
		size := formatSize(len(body))
		files = append(files, fileEntry{name: fname, size: size})
	}

	// Build Apache-style HTML
	serverVersion := m.ServerVersion
	if serverVersion == "" {
		serverVersion = "Apache/2.4.41 (Ubuntu)"
	}
	serverName := m.ServerName
	if serverName == "" {
		serverName = "localhost"
	}

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 3.2 Final//EN\">\n")
	sb.WriteString("<html>\n <head>\n  <title>Index of ")
	sb.WriteString(htmlEscape(reqPath))
	sb.WriteString("</title>\n </head>\n <body>\n")
	sb.WriteString("<h1>Index of ")
	sb.WriteString(htmlEscape(reqPath))
	sb.WriteString("</h1>\n")
	sb.WriteString("  <table>\n")
	sb.WriteString("   <tr><th valign=\"top\"><img src=\"/icons/blank.svg\" alt=\"[ICO]\"></th>")
	sb.WriteString("<th><a href=\"?C=N;O=D\">Name</a></th>")
	sb.WriteString("<th><a href=\"?C=M;O=A\">Last modified</a></th>")
	sb.WriteString("<th><a href=\"?C=S;O=A\">Size</a></th>")
	sb.WriteString("<th><a href=\"?C=D;O=A\">Description</a></th></tr>\n")
	sb.WriteString("   <tr><th colspan=\"5\"><hr></th></tr>\n")

	// Parent directory link
	if reqPath != "/" {
		parent := path.Dir(reqPath)
		if parent == "." {
			parent = "/"
		}
		sb.WriteString(fmt.Sprintf("   <tr><td valign=\"top\"><img src=\"/icons/back.svg\" alt=\"[PARENTDIR]\"></td><td><a href=\"%s\">Parent Directory</a></td><td>&nbsp;</td><td align=\"right\">  - </td><td>&nbsp;</td></tr>\n", htmlEscape(parent)))
	}

	// Base time for modification dates, deterministic per path
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(r.Int63n(int64(365*24) * int64(time.Hour))))

	// Directories
	for _, d := range dirs {
		modTime := baseTime.Add(time.Duration(r.Int63n(int64(180*24) * int64(time.Hour))))
		href := path.Join(reqPath, d) + "/"
		sb.WriteString(fmt.Sprintf("   <tr><td valign=\"top\"><img src=\"/icons/folder.svg\" alt=\"[DIR]\"></td><td><a href=\"%s\">%s/</a></td><td align=\"right\">%s  </td><td align=\"right\">  - </td><td>&nbsp;</td></tr>\n",
			htmlEscape(href), htmlEscape(d), modTime.Format("2006-01-02 15:04")))
	}

	// Files
	for _, f := range files {
		modTime := baseTime.Add(time.Duration(r.Int63n(int64(180*24) * int64(time.Hour))))
		href := path.Join(reqPath, f.name)
		icon := iconForFile(f.name)
		sb.WriteString(fmt.Sprintf("   <tr><td valign=\"top\"><img src=\"/icons/%s\" alt=\"[   ]\"></td><td><a href=\"%s\">%s</a></td><td align=\"right\">%s  </td><td align=\"right\">%s</td><td>&nbsp;</td></tr>\n",
			icon, htmlEscape(href), htmlEscape(f.name), modTime.Format("2006-01-02 15:04"), f.size))
	}

	sb.WriteString("   <tr><th colspan=\"5\"><hr></th></tr>\n")
	sb.WriteString("  </table>\n")
	sb.WriteString(fmt.Sprintf("<address>%s Server at %s Port 80</address>\n", htmlEscape(serverVersion), htmlEscape(serverName)))
	sb.WriteString("</body></html>\n")

	return MazeResponse{
		StatusCode:  http.StatusOK,
		ContentType: "text/html; charset=UTF-8",
		Body:        sb.String(),
		Headers: map[string]string{
			"Server": serverVersion,
		},
	}
}

func (m *MazeHoneypot) generateFileResponse(reqPath string) MazeResponse {
	r := rand.New(rand.NewSource(seedFromPath(reqPath)))
	base := path.Base(reqPath)

	// Find matching template by extension or name
	var genFunc func(r *rand.Rand, fullPath string) string
	contentType := "text/plain; charset=UTF-8"

	for _, tmpl := range allFileTemplates {
		if base == tmpl.name+tmpl.ext {
			genFunc = tmpl.genFunc
			break
		}
	}
	if genFunc == nil {
		for _, tmpl := range allFileTemplates {
			if tmpl.ext != "" && path.Ext(base) == tmpl.ext {
				genFunc = tmpl.genFunc
				break
			}
		}
	}

	// Match by name prefix for dotfiles
	if genFunc == nil {
		switch {
		case base == ".env":
			genFunc = genEnvFile
		case base == ".htaccess":
			genFunc = genHtaccess
		case base == ".gitignore":
			genFunc = genGitignore
		case base == "Makefile":
			genFunc = genMakefile
		case base == "Dockerfile":
			genFunc = genDockerfile
		case strings.HasSuffix(base, ".sql"):
			genFunc = genSQLDump
		case strings.HasSuffix(base, ".log"):
			genFunc = genAccessLog
		case strings.HasSuffix(base, ".php"):
			genFunc = genPHPIndex
		case strings.HasSuffix(base, ".py"):
			genFunc = genPythonApp
		case strings.HasSuffix(base, ".go"):
			genFunc = genGoMain
		case strings.HasSuffix(base, ".sh"):
			genFunc = genDeployScript
		case strings.HasSuffix(base, ".yaml"), strings.HasSuffix(base, ".yml"):
			genFunc = genYAMLConfig
		case strings.HasSuffix(base, ".json"):
			genFunc = genJSONConfig
		case strings.HasSuffix(base, ".conf"):
			genFunc = genNginxConf
		case strings.HasSuffix(base, ".csv"):
			genFunc = genUsersCSV
		case strings.HasSuffix(base, ".md"):
			genFunc = genReadme
		case strings.HasSuffix(base, ".txt"):
			genFunc = genCredentialsTxt
		case strings.HasSuffix(base, ".tar.gz"), strings.HasSuffix(base, ".zip"),
			strings.HasSuffix(base, ".key"), strings.HasSuffix(base, ".pem"):
			genFunc = genBinaryPlaceholder
		default:
			genFunc = genGenericFile
		}
	}

	// Set content type based on extension
	ext := path.Ext(base)
	switch ext {
	case ".html", ".htm":
		contentType = "text/html; charset=UTF-8"
	case ".json":
		contentType = "application/json"
	case ".yaml", ".yml":
		contentType = "text/yaml; charset=UTF-8"
	case ".php":
		contentType = "text/html; charset=UTF-8"
	case ".csv":
		contentType = "text/csv; charset=UTF-8"
	case ".sql":
		contentType = "application/sql"
	case ".sh":
		contentType = "application/x-sh"
	case ".py":
		contentType = "text/x-python"
	case ".go":
		contentType = "text/x-go"
	case ".tar.gz":
		contentType = "application/gzip"
	case ".zip":
		contentType = "application/zip"
	case ".key", ".pem":
		contentType = "application/x-pem-file"
	case ".md":
		contentType = "text/markdown; charset=UTF-8"
	case ".xml":
		contentType = "application/xml"
	}

	body := genFunc(r, reqPath)

	serverVersion := m.ServerVersion
	if serverVersion == "" {
		serverVersion = "Apache/2.4.41 (Ubuntu)"
	}

	return MazeResponse{
		StatusCode:  http.StatusOK,
		ContentType: contentType,
		Body:        body,
		Headers: map[string]string{
			"Server":         serverVersion,
			"Content-Length": fmt.Sprintf("%d", len(body)),
		},
	}
}

// --- File content generators ---

func genSQLDump(r *rand.Rand, fullPath string) string {
	dbName := pickOne(r, []string{"production_db", "webapp_db", "customers", "main_app", "ecommerce", "analytics"})
	tables := []string{"users", "orders", "sessions", "payments", "products", "audit_log"}
	r.Shuffle(len(tables), func(i, j int) { tables[i], tables[j] = tables[j], tables[i] })

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("-- MySQL dump 10.13  Distrib 8.0.%d\n", 28+r.Intn(8)))
	sb.WriteString(fmt.Sprintf("-- Host: db-master-%d.internal    Database: %s\n", r.Intn(10), dbName))
	sb.WriteString("-- Server version\t8.0.32\n\n")
	sb.WriteString("/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;\n")
	sb.WriteString("/*!40101 SET NAMES utf8mb4 */;\n\n")

	for i := 0; i < 2+r.Intn(2); i++ {
		table := tables[i%len(tables)]
		sb.WriteString(fmt.Sprintf("--\n-- Table structure for table `%s`\n--\n\n", table))
		sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", table))
		sb.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", table))
		sb.WriteString("  `id` bigint unsigned NOT NULL AUTO_INCREMENT,\n")
		if table == "users" {
			sb.WriteString("  `email` varchar(255) NOT NULL,\n")
			sb.WriteString("  `password_hash` varchar(255) NOT NULL,\n")
			sb.WriteString("  `role` enum('admin','user','moderator') DEFAULT 'user',\n")
			sb.WriteString("  `api_key` varchar(64) DEFAULT NULL,\n")
		} else if table == "orders" {
			sb.WriteString("  `user_id` bigint unsigned NOT NULL,\n")
			sb.WriteString("  `total_amount` decimal(10,2) NOT NULL,\n")
			sb.WriteString("  `status` enum('pending','shipped','delivered') DEFAULT 'pending',\n")
		} else if table == "payments" {
			sb.WriteString("  `order_id` bigint unsigned NOT NULL,\n")
			sb.WriteString("  `stripe_token` varchar(128) DEFAULT NULL,\n")
			sb.WriteString("  `amount` decimal(10,2) NOT NULL,\n")
		} else {
			sb.WriteString("  `name` varchar(255) DEFAULT NULL,\n")
			sb.WriteString("  `data` text,\n")
		}
		sb.WriteString("  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,\n")
		sb.WriteString("  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n")
		sb.WriteString(fmt.Sprintf("  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 AUTO_INCREMENT=%d;\n\n", 1000+r.Intn(50000)))

		// Insert some fake rows
		sb.WriteString(fmt.Sprintf("--\n-- Dumping data for table `%s`\n--\n\n", table))
		sb.WriteString(fmt.Sprintf("LOCK TABLES `%s` WRITE;\n", table))
		if table == "users" {
			for j := 0; j < 3+r.Intn(5); j++ {
				user := pickOne(r, []string{"admin", "john.doe", "jane.smith", "dev_ops", "backup_admin", "service_account", "deploy_bot", "root"})
				domain := pickOne(r, []string{"company.com", "internal.io", "corp.net", "example.org"})
				sb.WriteString(fmt.Sprintf("INSERT INTO `%s` VALUES (%d,'%s@%s','$2b$12$%s','%s','ak_%s','2023-%02d-%02d 10:%02d:00',NULL);\n",
					table, j+1, user, domain, randomHash(r, 53), pickOne(r, []string{"admin", "user", "user"}),
					randomHex(r, 16), 1+r.Intn(12), 1+r.Intn(28), r.Intn(24)))
			}
		}
		sb.WriteString(fmt.Sprintf("UNLOCK TABLES;\n\n"))
	}

	sb.WriteString("/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;\n")
	return sb.String()
}

func genYAMLConfig(r *rand.Rand, fullPath string) string {
	appName := pickOne(r, []string{"webapp", "api-server", "microservice", "backend", "worker"})
	return fmt.Sprintf(`# Application configuration
app:
  name: %s
  version: %d.%d.%d
  environment: production
  debug: false

server:
  host: 0.0.0.0
  port: %d
  workers: %d

database:
  host: db-master-%d.internal
  port: 5432
  name: %s_production
  username: app_user
  password: %s
  pool_size: %d
  ssl_mode: require

redis:
  host: redis-%d.internal
  port: 6379
  password: %s
  db: 0

aws:
  region: %s
  access_key_id: AKIA%s
  secret_access_key: %s
  s3_bucket: %s-assets-prod

logging:
  level: info
  format: json
  output: /var/log/%s/app.log
`,
		appName, 1+r.Intn(4), r.Intn(10), r.Intn(20),
		3000+r.Intn(6000), 2+r.Intn(16),
		r.Intn(5), appName,
		randomAlphaNum(r, 24), 5+r.Intn(20),
		r.Intn(3), randomAlphaNum(r, 16),
		pickOne(r, []string{"us-east-1", "eu-west-1", "ap-southeast-1", "us-west-2"}),
		randomAlphaUpper(r, 16), randomBase64(r, 40),
		appName, appName)
}

func genJSONConfig(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "version": "%d.%d.%d",
  "database": {
    "host": "db-%d.internal.%s",
    "port": 5432,
    "username": "app_service",
    "password": "%s"
  },
  "jwt_secret": "%s",
  "api_keys": {
    "stripe": "sk_live_%s",
    "sendgrid": "SG.%s",
    "twilio_sid": "AC%s"
  },
  "admin_emails": [
    "admin@%s",
    "devops@%s"
  ]
}`,
		pickOne(r, []string{"api-gateway", "auth-service", "payment-service", "user-service"}),
		1+r.Intn(3), r.Intn(10), r.Intn(30),
		r.Intn(5), pickOne(r, []string{"company.com", "corp.io", "example.net"}),
		randomAlphaNum(r, 20),
		randomBase64(r, 32),
		randomAlphaNum(r, 24),
		randomBase64(r, 22),
		randomHex(r, 32),
		pickOne(r, []string{"acme-corp.com", "widgets.io", "startup.dev"}),
		pickOne(r, []string{"acme-corp.com", "widgets.io", "startup.dev"}))
}

func genEnvFile(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`# Environment configuration - DO NOT COMMIT
NODE_ENV=production
PORT=%d
DATABASE_URL=postgresql://app:%s@db-%d.internal:5432/production
REDIS_URL=redis://:%s@redis-%d.internal:6379/0
JWT_SECRET=%s
SESSION_SECRET=%s

# AWS Credentials
AWS_ACCESS_KEY_ID=AKIA%s
AWS_SECRET_ACCESS_KEY=%s
AWS_REGION=%s
S3_BUCKET=prod-assets-%s

# API Keys
STRIPE_SECRET_KEY=sk_live_%s
STRIPE_WEBHOOK_SECRET=whsec_%s
SENDGRID_API_KEY=SG.%s
GITHUB_TOKEN=ghp_%s

# Internal services
AUTH_SERVICE_URL=http://auth.internal:3001
PAYMENT_SERVICE_URL=http://payments.internal:3002
ADMIN_API_KEY=%s
`,
		3000+r.Intn(6000),
		randomAlphaNum(r, 20), r.Intn(5),
		randomAlphaNum(r, 16), r.Intn(3),
		randomBase64(r, 32),
		randomBase64(r, 32),
		randomAlphaUpper(r, 16),
		randomBase64(r, 40),
		pickOne(r, []string{"us-east-1", "eu-west-1", "us-west-2", "ap-southeast-1"}),
		randomHex(r, 8),
		randomAlphaNum(r, 24),
		randomAlphaNum(r, 24),
		randomBase64(r, 22),
		randomAlphaNum(r, 36),
		randomHex(r, 32))
}

func genAccessLog(r *rand.Rand, fullPath string) string {
	var sb strings.Builder
	ips := []string{
		fmt.Sprintf("192.168.%d.%d", r.Intn(256), 1+r.Intn(254)),
		fmt.Sprintf("10.0.%d.%d", r.Intn(256), 1+r.Intn(254)),
		fmt.Sprintf("%d.%d.%d.%d", 40+r.Intn(200), r.Intn(256), r.Intn(256), 1+r.Intn(254)),
	}
	paths := []string{"/", "/admin", "/api/v1/users", "/login", "/dashboard", "/api/v1/config",
		"/static/js/app.js", "/wp-admin/", "/phpmyadmin/", "/.env", "/api/v1/export"}
	methods := []string{"GET", "GET", "GET", "POST", "GET", "GET", "GET", "GET", "GET", "GET", "GET"}
	codes := []int{200, 200, 200, 302, 200, 403, 200, 301, 200, 200, 200}

	baseDate := time.Date(2024, time.Month(1+r.Intn(12)), 1+r.Intn(28), 0, 0, 0, 0, time.UTC)
	for i := 0; i < 20+r.Intn(30); i++ {
		ip := ips[r.Intn(len(ips))]
		pIdx := r.Intn(len(paths))
		t := baseDate.Add(time.Duration(i) * time.Duration(30+r.Intn(300)) * time.Second)
		size := 200 + r.Intn(50000)
		ua := pickOne(r, []string{
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"curl/7.68.0",
			"python-requests/2.28.1",
			"Go-http-client/1.1",
		})
		sb.WriteString(fmt.Sprintf("%s - - [%s] \"%s %s HTTP/1.1\" %d %d \"-\" \"%s\"\n",
			ip, t.Format("02/Jan/2006:15:04:05 -0700"), methods[pIdx], paths[pIdx], codes[pIdx], size, ua))
	}
	return sb.String()
}

func genErrorLog(r *rand.Rand, fullPath string) string {
	var sb strings.Builder
	baseDate := time.Date(2024, time.Month(1+r.Intn(12)), 1+r.Intn(28), 0, 0, 0, 0, time.UTC)
	errors := []string{
		"[error] [pid %d] [client %s] File does not exist: /var/www/html/robots.txt",
		"[warn] [pid %d] [client %s] ModSecurity: Access denied with code 403",
		"[error] [pid %d] [client %s] PHP Fatal error: Uncaught Error: Call to undefined function mysql_connect()",
		"[error] [pid %d] [client %s] AH01630: client denied by server configuration: /var/www/html/.git",
		"[warn] [pid %d] [client %s] mod_fcgid: stderr: PHP Warning: file_get_contents(/tmp/sess_) failed",
		"[error] [pid %d] [client %s] script '/var/www/html/wp-login.php' not found or unable to stat",
	}
	for i := 0; i < 10+r.Intn(15); i++ {
		t := baseDate.Add(time.Duration(i) * time.Duration(60+r.Intn(600)) * time.Second)
		ip := fmt.Sprintf("%d.%d.%d.%d", 40+r.Intn(200), r.Intn(256), r.Intn(256), 1+r.Intn(254))
		msg := fmt.Sprintf(errors[r.Intn(len(errors))], 1000+r.Intn(30000), ip)
		sb.WriteString(fmt.Sprintf("[%s] %s\n", t.Format("Mon Jan 02 15:04:05.000000 2006"), msg))
	}
	return sb.String()
}

func genCredentialsTxt(r *rand.Rand, fullPath string) string {
	var sb strings.Builder
	sb.WriteString("# Service credentials - INTERNAL USE ONLY\n")
	sb.WriteString("# Last updated by ops team\n\n")
	services := []string{"MySQL", "PostgreSQL", "Redis", "MongoDB", "RabbitMQ", "Elasticsearch", "Jenkins", "Grafana"}
	for _, svc := range services {
		if r.Intn(3) == 0 {
			continue
		}
		user := pickOne(r, []string{"admin", "root", "service", "app_user", "deployer"})
		sb.WriteString(fmt.Sprintf("## %s\nHost: %s-%d.internal\nUser: %s\nPass: %s\n\n",
			svc, strings.ToLower(svc), r.Intn(5), user, randomAlphaNum(r, 16+r.Intn(8))))
	}
	return sb.String()
}

func genDeployScript(r *rand.Rand, fullPath string) string {
	env := pickOne(r, []string{"production", "staging", "canary"})
	return fmt.Sprintf(`#!/bin/bash
set -euo pipefail

# Deployment script for %s
# Usage: ./deploy.sh [version]

DEPLOY_ENV="%s"
APP_NAME="%s"
DOCKER_REGISTRY="registry.internal:%d"
CLUSTER="k8s-%s-%d"
NAMESPACE="${APP_NAME}-${DEPLOY_ENV}"

VERSION="${1:-latest}"

echo "[$(date)] Starting deployment of ${APP_NAME}:${VERSION} to ${DEPLOY_ENV}..."

# Authenticate with registry
docker login "${DOCKER_REGISTRY}" -u deploy -p "%s"

# Pull latest image
docker pull "${DOCKER_REGISTRY}/${APP_NAME}:${VERSION}"

# Run database migrations
echo "[$(date)] Running migrations..."
kubectl -n "${NAMESPACE}" exec deploy/${APP_NAME} -- ./manage.py migrate --no-input

# Rolling update
kubectl -n "${NAMESPACE}" set image deployment/${APP_NAME} \
  app="${DOCKER_REGISTRY}/${APP_NAME}:${VERSION}"

kubectl -n "${NAMESPACE}" rollout status deployment/${APP_NAME} --timeout=300s

echo "[$(date)] Deployment complete."

# Notify Slack
curl -s -X POST "%s" \
  -H 'Content-type: application/json' \
  -d "{\"text\":\"Deployed ${APP_NAME}:${VERSION} to ${DEPLOY_ENV}\"}"
`,
		env, env,
		pickOne(r, []string{"api-gateway", "web-app", "worker", "auth-service"}),
		5000+r.Intn(1000),
		env, 1+r.Intn(5),
		randomAlphaNum(r, 20),
		fmt.Sprintf("https://hooks.slack.com/services/T%s/B%s/%s",
			randomAlphaUpper(r, 8), randomAlphaUpper(r, 8), randomAlphaNum(r, 24)))
}

func genMakefile(r *rand.Rand, fullPath string) string {
	app := pickOne(r, []string{"app", "server", "api", "service"})
	return fmt.Sprintf(`.PHONY: build test deploy clean

APP_NAME := %s
VERSION := $(shell git describe --tags --always)
DOCKER_REGISTRY := registry.internal:%d

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	go test ./... -v -cover -race

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION) .

docker-push: docker-build
	docker push $(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION)

deploy-staging:
	kubectl -n staging set image deployment/$(APP_NAME) app=$(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION)

deploy-production:
	@echo "Deploying $(VERSION) to production..."
	kubectl -n production set image deployment/$(APP_NAME) app=$(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION)

migrate:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate

clean:
	rm -rf bin/ dist/
`, app, 5000+r.Intn(1000))
}

func genDockerCompose(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`version: '3.8'

services:
  app:
    build: .
    ports:
      - "%d:8080"
    environment:
      - DATABASE_URL=postgresql://app:%s@db:5432/production
      - REDIS_URL=redis://redis:6379/0
      - JWT_SECRET=%s
    depends_on:
      - db
      - redis
    restart: unless-stopped

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: production
      POSTGRES_USER: app
      POSTGRES_PASSWORD: %s
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass %s
    ports:
      - "6379:6379"

  nginx:
    image: nginx:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./certs:/etc/nginx/certs

volumes:
  pgdata:
`, 8080+r.Intn(100),
		randomAlphaNum(r, 16),
		randomBase64(r, 32),
		randomAlphaNum(r, 16),
		randomAlphaNum(r, 16))
}

func genReadme(r *rand.Rand, fullPath string) string {
	project := pickOne(r, []string{"InternalAPI", "CustomerPortal", "DataPipeline", "AdminDashboard", "PaymentGateway"})
	return fmt.Sprintf(`# %s

Internal service for %s operations.

## Quick Start

%ssh
cp .env.example .env
docker-compose up -d
make migrate
make run
%s

## Configuration

See %s.env%s for required environment variables.

## Deployment

Deployments are handled via Jenkins pipeline. See %sDeploy Guide%s.

## Team

- Backend: @backend-team
- DevOps: @ops
- Security: @security-team
`, project,
		pickOne(r, []string{"customer management", "payment processing", "data analytics", "user authentication"}),
		"```", "```", "`", "`", "[", "](./docs/deploy.md)")
}

func genPHPIndex(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`<?php
/**
 * Main entry point
 * @version %d.%d.%d
 */

require_once __DIR__ . '/vendor/autoload.php';

$config = require __DIR__ . '/config/database.php';

// Database connection
$dsn = sprintf("mysql:host=%%s;dbname=%%s;charset=utf8mb4",
    $config['host'] ?? 'db-%d.internal',
    $config['database'] ?? 'production'
);

try {
    $pdo = new PDO($dsn, $config['username'], $config['password'], [
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
    ]);
} catch (PDOException $e) {
    error_log("Database connection failed: " . $e->getMessage());
    http_response_code(500);
    exit("Internal Server Error");
}

$router = new Router();
$router->get('/', 'HomeController@index');
$router->get('/admin', 'AdminController@dashboard');
$router->post('/api/login', 'AuthController@login');
$router->dispatch($_SERVER['REQUEST_URI']);
`, 1+r.Intn(4), r.Intn(10), r.Intn(20), r.Intn(5))
}

func genWPConfig(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`<?php
/**
 * WordPress configuration
 * Generated: %d-%02d-%02d
 */

define('DB_NAME', 'wordpress_%s');
define('DB_USER', '%s');
define('DB_PASSWORD', '%s');
define('DB_HOST', 'db-%d.internal');
define('DB_CHARSET', 'utf8mb4');

define('AUTH_KEY',         '%s');
define('SECURE_AUTH_KEY',  '%s');
define('LOGGED_IN_KEY',    '%s');
define('NONCE_KEY',        '%s');

$table_prefix = 'wp_';
define('WP_DEBUG', false);
define('DISALLOW_FILE_EDIT', true);

if ( !defined('ABSPATH') )
    define('ABSPATH', dirname(__FILE__) . '/');

require_once(ABSPATH . 'wp-settings.php');
`,
		2023+r.Intn(3), 1+r.Intn(12), 1+r.Intn(28),
		pickOne(r, []string{"prod", "main", "live", "site"}),
		pickOne(r, []string{"wp_admin", "root", "db_user"}),
		randomAlphaNum(r, 20),
		r.Intn(5),
		randomBase64(r, 64),
		randomBase64(r, 64),
		randomBase64(r, 64),
		randomBase64(r, 64))
}

func genHtaccess(r *rand.Rand, fullPath string) string {
	return `# Apache configuration
RewriteEngine On
RewriteBase /

# Force HTTPS
RewriteCond %{HTTPS} off
RewriteRule ^(.*)$ https://%{HTTP_HOST}%{REQUEST_URI} [L,R=301]

# Block access to sensitive files
<FilesMatch "\.(env|git|sql|bak|old|log)$">
    Order allow,deny
    Deny from all
</FilesMatch>

# WordPress permalinks
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule . /index.php [L]

# Prevent directory listing
Options -Indexes

# Security headers
Header set X-Content-Type-Options "nosniff"
Header set X-Frame-Options "SAMEORIGIN"
Header set X-XSS-Protection "1; mode=block"
`
}

func genGitignore(r *rand.Rand, fullPath string) string {
	return `# Dependencies
node_modules/
vendor/
.venv/

# Environment
.env
.env.local
.env.production

# Build
dist/
build/
bin/

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Secrets
*.key
*.pem
credentials.json
service-account.json
`
}

func genSSHPubKey(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQ%s %s@%s\n",
		randomBase64(r, 200),
		pickOne(r, []string{"deploy", "admin", "ops", "service", "root"}),
		pickOne(r, []string{"prod-bastion", "jump-host", "deploy-server", "ci-runner"}))
}

func genUsersCSV(r *rand.Rand, fullPath string) string {
	var sb strings.Builder
	sb.WriteString("id,email,name,role,department,last_login\n")
	firstNames := []string{"James", "Sarah", "Michael", "Emily", "David", "Jessica", "Robert", "Amanda", "Daniel", "Maria"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Garcia", "Wilson", "Taylor"}
	roles := []string{"admin", "user", "manager", "developer", "analyst", "devops"}
	depts := []string{"Engineering", "Operations", "Finance", "Security", "HR", "Product"}
	domain := pickOne(r, []string{"company.com", "corp.io", "acme.org"})

	for i := 0; i < 10+r.Intn(15); i++ {
		first := firstNames[r.Intn(len(firstNames))]
		last := lastNames[r.Intn(len(lastNames))]
		email := fmt.Sprintf("%s.%s@%s", strings.ToLower(first), strings.ToLower(last), domain)
		sb.WriteString(fmt.Sprintf("%d,%s,%s %s,%s,%s,2024-%02d-%02d\n",
			1000+i, email, first, last, roles[r.Intn(len(roles))],
			depts[r.Intn(len(depts))], 1+r.Intn(12), 1+r.Intn(28)))
	}
	return sb.String()
}

func genRequirementsTxt(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`Django==%d.%d.%d
djangorestframework==3.%d.%d
psycopg2-binary==2.9.%d
redis==%d.%d.%d
celery==%d.%d.%d
gunicorn==21.%d.%d
boto3==1.%d.%d
requests==2.%d.%d
PyJWT==%d.%d.%d
cryptography==%d.%d.%d
python-dotenv==1.%d.%d
sentry-sdk==1.%d.%d
`,
		4+r.Intn(2), r.Intn(3), r.Intn(10),
		14, r.Intn(5),
		r.Intn(10),
		4+r.Intn(2), r.Intn(5), r.Intn(10),
		5, r.Intn(4), r.Intn(5),
		r.Intn(3), r.Intn(5),
		28+r.Intn(6), r.Intn(100),
		28+r.Intn(4), r.Intn(5),
		2, r.Intn(8), r.Intn(5),
		40+r.Intn(5), r.Intn(5), r.Intn(5),
		r.Intn(2), r.Intn(5),
		r.Intn(40), r.Intn(10))
}

func genPackageJSON(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "version": "%d.%d.%d",
  "private": true,
  "scripts": {
    "start": "node dist/index.js",
    "dev": "ts-node src/index.ts",
    "build": "tsc",
    "test": "jest --coverage",
    "migrate": "knex migrate:latest",
    "deploy": "npm run build && pm2 restart ecosystem.config.js"
  },
  "dependencies": {
    "express": "^4.18.%d",
    "pg": "^8.%d.%d",
    "redis": "^4.%d.%d",
    "jsonwebtoken": "^9.0.%d",
    "bcrypt": "^5.1.%d",
    "dotenv": "^16.%d.%d",
    "axios": "^1.%d.%d",
    "winston": "^3.%d.%d"
  },
  "devDependencies": {
    "typescript": "^5.%d.%d",
    "jest": "^29.%d.%d",
    "@types/node": "^20.%d.%d"
  }
}`,
		pickOne(r, []string{"api-server", "backend-service", "payment-gateway", "auth-service"}),
		1+r.Intn(3), r.Intn(10), r.Intn(20),
		r.Intn(5),
		r.Intn(12), r.Intn(10),
		r.Intn(7), r.Intn(10),
		r.Intn(5),
		r.Intn(5),
		r.Intn(5), r.Intn(10),
		r.Intn(7), r.Intn(10),
		r.Intn(12), r.Intn(10),
		r.Intn(5), r.Intn(5),
		r.Intn(7), r.Intn(10),
		r.Intn(12), r.Intn(20))
}

func genTerraformState(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`{
  "version": 4,
  "terraform_version": "1.%d.%d",
  "serial": %d,
  "lineage": "%s",
  "outputs": {
    "db_endpoint": {
      "value": "db-%d.%s.rds.amazonaws.com",
      "type": "string"
    },
    "api_url": {
      "value": "https://api.%s",
      "type": "string"
    }
  },
  "resources": [
    {
      "mode": "managed",
      "type": "aws_instance",
      "name": "web_server",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "attributes": {
            "id": "i-%s",
            "ami": "ami-%s",
            "instance_type": "%s",
            "vpc_security_group_ids": ["sg-%s"],
            "subnet_id": "subnet-%s",
            "key_name": "prod-deploy-key"
          }
        }
      ]
    }
  ]
}`,
		5+r.Intn(4), r.Intn(10),
		100+r.Intn(500),
		randomHex(r, 8)+"-"+randomHex(r, 4)+"-"+randomHex(r, 4)+"-"+randomHex(r, 4)+"-"+randomHex(r, 12),
		r.Intn(5),
		pickOne(r, []string{"us-east-1", "eu-west-1", "us-west-2"}),
		pickOne(r, []string{"acme-corp.com", "widgets.io", "startup.dev"}),
		randomHex(r, 17),
		randomHex(r, 8),
		pickOne(r, []string{"t3.large", "m5.xlarge", "c5.2xlarge", "r5.large"}),
		randomHex(r, 8),
		randomHex(r, 8))
}

func genDockerfile(r *rand.Rand, fullPath string) string {
	lang := r.Intn(3)
	switch lang {
	case 0:
		return fmt.Sprintf(`FROM golang:1.%d-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

FROM alpine:3.%d
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config
EXPOSE %d
CMD ["./server"]
`, 20+r.Intn(4), 17+r.Intn(3), 3000+r.Intn(6000))
	case 1:
		return fmt.Sprintf(`FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json .
EXPOSE %d
USER node
CMD ["node", "dist/index.js"]
`, 3000+r.Intn(6000))
	default:
		return fmt.Sprintf(`FROM python:3.%d-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE %d
CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:%d", "app:create_app()"]
`, 10+r.Intn(3), 8000+r.Intn(2000), 8000+r.Intn(2000))
	}
}

func genNginxConf(r *rand.Rand, fullPath string) string {
	domain := pickOne(r, []string{"api.company.com", "app.internal.io", "dashboard.corp.net"})
	return fmt.Sprintf(`upstream backend {
    server 127.0.0.1:%d;
    server 127.0.0.1:%d backup;
}

server {
    listen 80;
    server_name %s;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name %s;

    ssl_certificate /etc/nginx/certs/fullchain.pem;
    ssl_certificate_key /etc/nginx/certs/privkey.pem;

    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /static/ {
        alias /var/www/static/;
        expires 30d;
    }

    access_log /var/log/nginx/%s.access.log;
    error_log /var/log/nginx/%s.error.log;
}
`, 3000+r.Intn(6000), 3000+r.Intn(6000), domain, domain, domain, domain)
}

func genMigrationSQL(r *rand.Rand, fullPath string) string {
	ver := fmt.Sprintf("%04d", r.Intn(200))
	return fmt.Sprintf(`-- Migration %s: %s
-- Created: %d-%02d-%02d

BEGIN;

ALTER TABLE users ADD COLUMN mfa_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN mfa_secret VARCHAR(64);
ALTER TABLE users ADD COLUMN last_password_change TIMESTAMP;

CREATE TABLE api_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(128) NOT NULL UNIQUE,
    name VARCHAR(255),
    scopes TEXT[],
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_api_tokens_user_id ON api_tokens(user_id);
CREATE INDEX idx_api_tokens_token ON api_tokens(token);

INSERT INTO schema_migrations (version) VALUES ('%s');

COMMIT;
`, ver, pickOne(r, []string{"add_mfa_support", "create_api_tokens", "add_audit_log", "add_user_roles"}),
		2023+r.Intn(3), 1+r.Intn(12), 1+r.Intn(28), ver)
}

func genTodoTxt(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`TODO - Sprint %d
================

[x] Fix authentication bypass in /api/admin endpoint
[x] Update database credentials (rotated after incident)
[ ] Remove hardcoded AWS keys from config.yaml
[ ] Enable WAF rules for SQL injection protection
[ ] Rotate JWT signing key (current key: %s)
[ ] Migrate from MD5 to bcrypt for password hashing
[ ] Fix CORS configuration allowing wildcard origins
[ ] Disable directory listing on Apache
[ ] Review IAM roles - service account has admin access
[ ] Update SSL certificate (expires %d-%02d-%02d)

CRITICAL:
- The staging database dump is still accessible at /backup/staging_dump.sql
- Admin panel has no rate limiting on login endpoint
- Legacy API endpoint /api/v1/debug still returns stack traces
`, r.Intn(50),
		randomBase64(r, 24),
		2024+r.Intn(2), 1+r.Intn(12), 1+r.Intn(28))
}

func genNotesMd(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`# Meeting Notes - %d-%02d-%02d

## Attendees
- @dev-lead, @ops, @security

## Discussion

### Infrastructure
- Migrate remaining services to k8s cluster %s
- Current DB credentials need rotation (last changed %d days ago)
- AWS access keys for CI/CD: AKIA%s (needs scoping down)

### Security Review
- Penetration test scheduled for next sprint
- Found exposed .git directory on production server
- SSH keys for bastion host should be rotated monthly
- API rate limiting not enforced on internal endpoints

### Action Items
1. [ ] Rotate all database passwords by EOW
2. [ ] Enable audit logging on production database
3. [ ] Review firewall rules for staging environment
4. [ ] Update incident response playbook
`,
		2024+r.Intn(2), 1+r.Intn(12), 1+r.Intn(28),
		pickOne(r, []string{"prod-east-1", "staging-west-2", "prod-eu-1"}),
		30+r.Intn(365),
		randomAlphaUpper(r, 16))
}

func genPythonApp(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`#!/usr/bin/env python3
"""Main application entry point."""

import os
from flask import Flask, jsonify, request
from functools import wraps
import jwt
import psycopg2

app = Flask(__name__)

# Configuration
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://app:%s@db-%d.internal:5432/production")
JWT_SECRET = os.getenv("JWT_SECRET", "%s")
ADMIN_API_KEY = os.getenv("ADMIN_API_KEY", "%s")

def get_db():
    return psycopg2.connect(DATABASE_URL)

def require_auth(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        token = request.headers.get("Authorization", "").replace("Bearer ", "")
        try:
            payload = jwt.decode(token, JWT_SECRET, algorithms=["HS256"])
            request.user = payload
        except jwt.InvalidTokenError:
            return jsonify({"error": "Invalid token"}), 401
        return f(*args, **kwargs)
    return decorated

@app.route("/api/v1/health")
def health():
    return jsonify({"status": "ok"})

@app.route("/api/v1/users")
@require_auth
def list_users():
    conn = get_db()
    cur = conn.cursor()
    cur.execute("SELECT id, email, role FROM users LIMIT 100")
    users = [{"id": r[0], "email": r[1], "role": r[2]} for r in cur.fetchall()]
    cur.close()
    conn.close()
    return jsonify(users)

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=%d, debug=False)
`,
		randomAlphaNum(r, 16), r.Intn(5),
		randomBase64(r, 32),
		randomHex(r, 32),
		5000+r.Intn(4000))
}

func genGoMain(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf(`package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var (
	dbURL     = envOr("DATABASE_URL", "postgres://app:%s@db-%d.internal:5432/production?sslmode=require")
	jwtSecret = envOr("JWT_SECRET", "%s")
	apiPort   = envOr("PORT", "%d")
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/api/v1/config", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"database": dbURL,
			"version":  "%d.%d.%d",
		})
	})

	log.Printf("Server starting on %%s", apiPort)
	log.Fatal(http.ListenAndServe(":"+apiPort, nil))
}
`,
		randomAlphaNum(r, 16), r.Intn(5),
		randomBase64(r, 32),
		3000+r.Intn(6000),
		1+r.Intn(3), r.Intn(10), r.Intn(20))
}

func genBinaryPlaceholder(r *rand.Rand, fullPath string) string {
	// Return a small amount of random-looking binary data header
	var sb strings.Builder
	for i := 0; i < 256; i++ {
		sb.WriteByte(byte(r.Intn(256)))
	}
	return sb.String()
}

func genGenericFile(r *rand.Rand, fullPath string) string {
	return fmt.Sprintf("# Auto-generated file\n# Path: %s\n# Last modified: %d-%02d-%02d\n\n%s\n",
		fullPath, 2023+r.Intn(3), 1+r.Intn(12), 1+r.Intn(28),
		pickOne(r, []string{
			"Configuration data placeholder",
			"Internal service documentation",
			"Legacy migration notes",
		}))
}

// --- Utility functions ---

func pickOne(r *rand.Rand, options []string) string {
	return options[r.Intn(len(options))]
}

func randomHex(r *rand.Rand, length int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, length)
	for i := range b {
		b[i] = hex[r.Intn(len(hex))]
	}
	return string(b)
}

func randomHash(r *rand.Rand, length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789./"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func randomAlphaNum(r *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func randomAlphaUpper(r *rand.Rand, length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func randomBase64(r *rand.Rand, length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func formatSize(bytes int) string {
	switch {
	case bytes >= 1048576:
		return fmt.Sprintf("%.1fM", float64(bytes)/1048576.0)
	case bytes >= 1024:
		return fmt.Sprintf("%.1fK", float64(bytes)/1024.0)
	default:
		return fmt.Sprintf("%d", bytes)
	}
}

func iconForFile(name string) string {
	ext := strings.ToLower(path.Ext(name))
	switch ext {
	case ".txt", ".md", ".log", ".csv":
		return "text.svg"
	case ".gz", ".tar", ".zip", ".bak":
		return "compressed.svg"
	case ".sh", ".py", ".go", ".php", ".js":
		return "script.svg"
	case ".yaml", ".yml", ".json", ".xml", ".conf", ".toml":
		return "layout.svg"
	case ".sql":
		return "layout.svg"
	case ".key", ".pem", ".pub":
		return "key.svg"
	default:
		return "unknown.svg"
	}
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
