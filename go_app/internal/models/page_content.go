package models

import "time"

// PageContent stores structured CMS JSON for editable public pages in SQLite.
type PageContent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PageKey   string    `gorm:"column:page_key;size:50;uniqueIndex;not null" json:"page_key"`
	DataJSON  string    `gorm:"column:data_json;type:text;not null" json:"data_json"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PageContent) TableName() string {
	return "page_contents"
}

// ============================================================
// SHARED SUB-TYPES
// ============================================================

type StatItem struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ============================================================
// 1. HOMEPAGE CONTENT (index.html)
// ============================================================

type HomePillar struct {
	Icon        string   `json:"icon"` // monitor, shield, zap, cpu, database, radio
	Tag         string   `json:"tag"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Chips       []string `json:"chips"`
	LinkText    string   `json:"link_text"`
	LinkURL     string   `json:"link_url"`
}

type HomeDemoCard struct {
	Tag         string   `json:"tag"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Chips       []string `json:"chips"`
	LinkText    string   `json:"link_text"`
	LinkURL     string   `json:"link_url"`
}

type HomeSecurityCard struct {
	Tag         string   `json:"tag"`
	RepoLabel   string   `json:"repo_label"`
	RepoURL     string   `json:"repo_url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Chips       []string `json:"chips"`
	LinkText    string   `json:"link_text"`
	LinkURL     string   `json:"link_url"`
}

type HomeProofRow struct {
	Icon        string `json:"icon"` // zap, shield, cpu, check
	Title       string `json:"title"`
	Description string `json:"description"`
}

type HomeContent struct {
	MetaTitle             string             `json:"meta_title"`
	MetaDescription       string             `json:"meta_description"`
	HeroEyebrow           string             `json:"hero_eyebrow"`
	HeroTitleLine1        string             `json:"hero_title_line1"`
	HeroTitleAccent       string             `json:"hero_title_accent"`
	HeroLead              string             `json:"hero_lead"`
	HeroChips             []string           `json:"hero_chips"`
	HeroPrimaryLabel      string             `json:"hero_primary_label"`
	HeroPrimaryURL        string             `json:"hero_primary_url"`
	HeroSecondaryLabel    string             `json:"hero_secondary_label"`
	HeroSecondaryURL      string             `json:"hero_secondary_url"`
	PillarsEyebrow        string             `json:"pillars_eyebrow"`
	PillarsTitleLine1     string             `json:"pillars_title_line1"`
	PillarsTitleAccent    string             `json:"pillars_title_accent"`
	PillarsLead           string             `json:"pillars_lead"`
	Pillars               []HomePillar       `json:"pillars"`
	DemosEyebrow          string             `json:"demos_eyebrow"`
	DemosTitle            string             `json:"demos_title"`
	DemosLead             string             `json:"demos_lead"`
	Demos                 []HomeDemoCard     `json:"demos"`
	SecurityEyebrow       string             `json:"security_eyebrow"`
	SecurityTitle         string             `json:"security_title"`
	SecurityLead          string             `json:"security_lead"`
	SecurityTools         []HomeSecurityCard `json:"security_tools"`
	PhilosophyEyebrow     string             `json:"philosophy_eyebrow"`
	PhilosophyTitleLine1  string             `json:"philosophy_title_line1"`
	PhilosophyTitleLine2  string             `json:"philosophy_title_line2"`
	PhilosophyTitleAccent string             `json:"philosophy_title_accent"`
	PhilosophyLead        string             `json:"philosophy_lead"`
	FaithTitle            string             `json:"faith_title"`
	FaithVerseRef         string             `json:"faith_verse_ref"`
	FaithVerseText        string             `json:"faith_verse_text"`
	FaithStatement        string             `json:"faith_statement"`
	ProofRows             []HomeProofRow     `json:"proof_rows"`
	CTABadge              string             `json:"cta_badge"`
	CTATitleLine1         string             `json:"cta_title_line1"`
	CTATitleLine2         string             `json:"cta_title_line2"`
	CTATitleAccent        string             `json:"cta_title_accent"`
	CTALead               string             `json:"cta_lead"`
}

func DefaultHomeContent() HomeContent {
	return HomeContent{
		MetaTitle:          "JO4 Dev • Software Engineer, AppSec Researcher & Freelance Systems Builder",
		MetaDescription:    "JO4 Dev by Jonathan Roets: High-performance Go Fiber backends, bespoke web apps, IoT telemetry, and offensive application security testing in South Africa.",
		HeroEyebrow:        "JO4 DEV • SOFTWARE ENGINEER • APPSEC • FREELANCE BUILDS",
		HeroTitleLine1:     "Software Engineer & Security Researcher.",
		HeroTitleAccent:    "Building & Breaking Systems.",
		HeroLead:           "Engineering high-performance Go Fiber systems, bespoke freelance web applications, enterprise automation, and developer-led offensive application security tooling.",
		HeroChips:          []string{"Go Fiber", "Freelance Builds", "AppSec • OWASP", "100k req/min APIs", "Hardware / IoT", "Sub-20ms TTFB"},
		HeroPrimaryLabel:   "Freelance Services & Demos",
		HeroPrimaryURL:     "/services",
		HeroSecondaryLabel: "Security Research",
		HeroSecondaryURL:   "/security",
		PillarsEyebrow:     "Three Core Pillars",
		PillarsTitleLine1:  "Full-Stack Freelance, AppSec & Distributed Systems.",
		PillarsTitleAccent: "Built for speed, security, and precision.",
		PillarsLead:        "Whether delivering custom production software for commercial clients, auditing complex web APIs for logic flaws, or optimizing compiled Go microservices for enterprise workloads.",
		Pillars: []HomePillar{
			{
				Icon:        "monitor",
				Tag:         "Pillar 01 • Freelance & Client Builds",
				Title:       "Freelance Web & Custom Software",
				Description: "Bespoke web applications, high-converting responsive business sites, ERP integrations, and workflow automation. Built with clean SSR, zero bloatware, sub-20ms response times, and 100/100 Lighthouse audits.",
				Chips:       []string{"Web Design", "Go Fiber SSR", "ERP Integrations", "SEO & Speed"},
				LinkText:    "Explore Freelance Services",
				LinkURL:     "/services",
			},
			{
				Icon:        "shield",
				Tag:         "Pillar 02 • Offensive Security & AppSec",
				Title:       "Security Tools & AppSec Audits",
				Description: "Custom CLI HTTP fuzzers, automated OWASP Top 10 test harnesses, and in-depth vulnerability research. Developer-led penetration testing that reveals authorization flaws (IDOR/BOLA) and business logic bypasses.",
				Chips:       []string{"Burp Suite", "Go Fuzzers", "OWASP Top 10", "IDOR & Logic Flaws"},
				LinkText:    "Explore Security Research",
				LinkURL:     "/security",
			},
			{
				Icon:        "zap",
				Tag:         "Pillar 03 • Distributed Systems & Hardware",
				Title:       "High-Throughput Systems Engineering",
				Description: "Enterprise packaging ERPs (PMD Online), computer vision die-line estimators, and low-power ESP32/LoRa RF livestock telemetry hardware. Built for microsecond query latency and zero fluff.",
				Chips:       []string{"Go (Fiber)", "OpenCV • Python", "MS SQL / Redis", "ESP32 • LoRa RF"},
				LinkText:    "View Systems Builds",
				LinkURL:     "/engineering",
			},
		},
		DemosEyebrow: "Live Demos & Builds",
		DemosTitle:   "Production Systems & Interactive Demos.",
		DemosLead:    "Real-world applications engineered for paying clients and production environments: from ERP manufacturing platforms to computer vision and IoT hardware.",
		Demos: []HomeDemoCard{
			{
				Tag:         "Enterprise ERP • Go Fiber",
				Title:       "PMD Online • Packaging ERP Platform",
				Description: "Comprehensive die-making and flexo plate ERP managing complex prepress tooling, rotary dies, and customer job queues with microsecond queries across MS SQL Server.",
				Chips:       []string{"Go Fiber", "MS SQL", "Redis", "Production ERP"},
				LinkText:    "View System Specs",
				LinkURL:     "/engineering#pmd-online",
			},
			{
				Tag:         "Computer Vision • Python",
				Title:       "Computer Vision Die-Line Estimator",
				Description: "Automated quoting engine that uses OpenCV contour extraction and vector geometry to measure knife lengths, crease rules, and perforation paths from DXF/PDF in under 400ms.",
				Chips:       []string{"OpenCV", "Python", "Geometric Quoting", "<400ms"},
				LinkText:    "View System Specs",
				LinkURL:     "/engineering#cv-quoting",
			},
			{
				Tag:         "IoT Hardware • Telemetry",
				Title:       "B.U.L.L. Agritech IoT Livestock Monitor",
				Description: "Solar-powered ESP32 collar integrating GPS and accelerometer telemetry transmitted over 915 MHz LoRa RF across rural cattle farms, feeding a high-concurrency time-series database.",
				Chips:       []string{"ESP32", "LoRa 915MHz", "C++ Firmware", "Agritech"},
				LinkText:    "View System Specs",
				LinkURL:     "/engineering#bull-agritech",
			},
		},
		SecurityEyebrow: "AppSec Research & Tools",
		SecurityTitle:   "Offensive Tooling & Vulnerability Research.",
		SecurityLead:    "Security tooling engineered in Go and Python to audit web APIs, simulate advanced adversary tactics, and automate AppSec validation.",
		SecurityTools: []HomeSecurityCard{
			{
				Tag:         "Offensive Tool • CLI",
				RepoLabel:   "GitHub",
				RepoURL:     "https://github.com/JRoetscyber",
				Title:       "Go-Fuzz-API • Concurrent HTTP Fuzzer",
				Description: "High-throughput API fuzzing engine built in Go using bounded worker pools and fasthttp. Discovers unauthenticated endpoints and rate-limit bypasses at over 10,000 req/sec.",
				Chips:       []string{"Go", "Goroutines", "API Fuzzing", "HTTP/2"},
				LinkText:    "Read Architecture",
				LinkURL:     "/security#go-fuzz-api",
			},
			{
				Tag:         "AppSec Research",
				RepoLabel:   "GitHub",
				RepoURL:     "https://github.com/JRoetscyber",
				Title:       "OWASP Top 10 Logic Flaw Hunter",
				Description: "Automated test suites designed to detect broken object-level authorization (BOLA/IDOR), multi-step workflow tampering, and concurrency race conditions in checkout APIs.",
				Chips:       []string{"Burp Suite", "IDOR Hunter", "Race Conditions", "Python"},
				LinkText:    "Read Case Study",
				LinkURL:     "/security#logic-flaw-hunter",
			},
			{
				Tag:         "Hardware • RF Security",
				RepoLabel:   "GitHub",
				RepoURL:     "https://github.com/JRoetscyber",
				Title:       "LoRa RF Telemetry Replay Toolkit",
				Description: "ESP32 transceiver firmware testing over-the-air IoT security. Captures packet frames, analyzes missing cryptographic nonces, and demonstrates replay mitigation on field devices.",
				Chips:       []string{"ESP32", "LoRa RF", "C++", "RF Testing"},
				LinkText:    "Read Teardown",
				LinkURL:     "/security#lora-rf-toolkit",
			},
		},
		PhilosophyEyebrow:     "Engineering Philosophy",
		PhilosophyTitleLine1:  "Constructing systems",
		PhilosophyTitleLine2:  "makes breaking them",
		PhilosophyTitleAccent: "second nature.",
		PhilosophyLead:        `Anyone can click "scan" on an automated scanner and generate a generic 50-page PDF report. Real-world application security requires understanding how the software was assembled: data flow bottlenecks, concurrency race conditions, database transactions, and authorization boundaries.`,
		FaithTitle:            "Faith-Driven Standards",
		FaithVerseRef:         "Colossians 3:23",
		FaithVerseText:        "Whatever you do, work at it with all your heart, as working for the Lord, not for human masters.",
		FaithStatement:        `"Work with purpose, build for His Glory." — Rigorous engineering ethics, total transparency, and unwavering integrity in every line of code and every penetration testing report.`,
		ProofRows: []HomeProofRow{
			{
				Icon:        "zap",
				Title:       "Compiled Go & Microsecond Latency",
				Description: "Go Fiber backends engineered with goroutines, singleflight Redis caching, and non-blocking batch flushing. Capable of handling over 100,000 requests per minute with sub-20ms response times.",
			},
			{
				Icon:        "shield",
				Title:       "Developer-Led AppSec & Logic Flaw Hunting",
				Description: "Uncovering the critical vulnerabilities automated tools miss: business logic flaws, broken object-level authorization (IDOR), SQL injection timing attacks, and authentication bypasses.",
			},
			{
				Icon:        "cpu",
				Title:       "Embedded Firmware & Telemetry",
				Description: "From low-level C++ firmware on ESP32 microcontrollers to 915 MHz LoRa RF packet parsing and GPS telemetry streams. Hands-on hardware reverse engineering and RF signal validation.",
			},
			{
				Icon:        "check",
				Title:       "Battle-Tested DevOps & Zero-Downtime Releases",
				Description: "Docker containerized environments, zero-downtime Blue/Green deployments, Nginx reverse proxy routing, and encrypted Cloudflare Tunnels for tamper-proof infrastructure.",
			},
		},
		CTABadge:       "Available for Freelance Builds, Security Audits & Engineering Roles",
		CTATitleLine1:  "Looking for a Full-Stack Freelancer,",
		CTATitleLine2:  "AppSec Researcher, or",
		CTATitleAccent: "Backend Systems Engineer?",
		CTALead:        "Whether you need a high-performance web application, an application security assessment of your web API, an automation system for your business, or are hiring for software engineering roles.",
	}
}

// ============================================================
// 2. ABOUT & SKILLS CONTENT (about.html)
// ============================================================

type AboutPrinciple struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type SkillCategory struct {
	Icon        string   `json:"icon"` // shield, zap, database, radio, code, monitor
	Tag         string   `json:"tag"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
}

type AboutContent struct {
	MetaTitle       string           `json:"meta_title"`
	MetaDescription string           `json:"meta_description"`
	HeroEyebrow     string           `json:"hero_eyebrow"`
	HeroTitle       string           `json:"hero_title"`
	HeroLead        string           `json:"hero_lead"`
	StatusBadge     string           `json:"status_badge"`
	AvatarImage     string           `json:"avatar_image"`
	FullName        string           `json:"full_name"`
	RoleTitle       string           `json:"role_title"`
	Location        string           `json:"location"`
	Email           string           `json:"email"`
	CVUrl           string           `json:"cv_url"`
	GitHubURL       string           `json:"github_url"`
	LinkedInURL     string           `json:"linkedin_url"`
	LeetCodeURL     string           `json:"leetcode_url"`
	BioEyebrow      string           `json:"bio_eyebrow"`
	BioTitleLine1   string           `json:"bio_title_line1"`
	BioTitleAccent  string           `json:"bio_title_accent"`
	BioParagraphs   []string         `json:"bio_paragraphs"`
	Principles      []AboutPrinciple `json:"principles"`
	SkillsEyebrow   string           `json:"skills_eyebrow"`
	SkillsTitle     string           `json:"skills_title"`
	SkillsLead      string           `json:"skills_lead"`
	SkillCategories []SkillCategory  `json:"skill_categories"`
	CTAEyebrow      string           `json:"cta_eyebrow"`
	CTATitleLine1   string           `json:"cta_title_line1"`
	CTATitleLine2   string           `json:"cta_title_line2"`
	CTALead         string           `json:"cta_lead"`
}

func DefaultAboutContent() AboutContent {
	return AboutContent{
		MetaTitle:       "About Jonathan Roets | Software & Security Engineer",
		MetaDescription: "Meet Jonathan Roets: Software Engineer & AppSec Researcher in South Africa specializing in high-concurrency Go backends and offensive penetration testing.",
		HeroEyebrow:     "PROFILE • BIOGRAPHY",
		HeroTitle:       "Building Systems. Breaking Flaws.",
		HeroLead:        "Software Engineer, AppSec Researcher, and Junior Penetration Tester based in Pretoria, South Africa.",
		StatusBadge:     "Available for AppSec, Pentesting & Backend Engineering Roles",
		AvatarImage:     "",
		FullName:        "Jonathan Roets",
		RoleTitle:       "Software & AppSec Engineer",
		Location:        "Pretoria, Gauteng, South Africa",
		Email:           "jroetscyber@gmail.com",
		CVUrl:           "/Jonathan_Roets_CV.pdf",
		GitHubURL:       "https://github.com/JRoetscyber",
		LinkedInURL:     "https://linkedin.com/in/jonathan-roets-b93420356",
		LeetCodeURL:     "https://leetcode.com/u/jroetscyber",
		BioEyebrow:      "The Evolution",
		BioTitleLine1:   "Why Builders Make The",
		BioTitleAccent:  "Most Dangerous Attackers.",
		BioParagraphs: []string{
			"My engineering journey didn't start with security theory or generic vulnerability scanners. It began in the trenches of production software development: building high-throughput ERP microservices in Go Fiber, engineering computer vision geometric estimation engines in Python and OpenCV, and writing bare-metal firmware for ESP32 and LoRa RF telemetry collars.",
			"When you spend years architecting distributed state machines, managing database isolation levels, preventing deadlocks, and writing high-concurrency goroutine workers, you develop a deep intuition for where systems break. You learn that vulnerabilities are rarely just missing input sanitization—the most devastating vulnerabilities are subtle business logic flaws, broken authorization contracts (IDOR), serialization desyncs, and state race conditions.",
			"This realization led naturally into offensive security research and penetration testing. Automated commercial scanners only find known CVE signatures; they are completely blind to business context. By approaching application security through the lens of a software architect, I can simulate sophisticated real-world adversaries who manipulate multi-step application states, bypass authorization models, and exploit the underlying system design.",
			"Today, I bridge both disciplines: engineering resilient, high-speed backend software, while conducting rigorous offensive penetration tests and authoring custom AppSec tooling to eliminate vulnerabilities before they reach production.",
		},
		Principles: []AboutPrinciple{
			{
				Title:       "Sub-20ms Standard",
				Description: "Zero bloatware, compiled Go, and cache-aside architectures with microsecond execution.",
			},
			{
				Title:       "Logic Over Scanners",
				Description: "Manual, developer-led penetration testing targeting architectural flaws scanners miss.",
			},
		},
		SkillsEyebrow: "Technical Competencies",
		SkillsTitle:   "Interactive Skills Matrix.",
		SkillsLead:    "Categorized expertise spanning offensive application security, distributed backend systems, production databases, and embedded IoT hardware.",
		SkillCategories: []SkillCategory{
			{
				Icon:        "shield",
				Tag:         "Primary Discipline",
				Title:       "Offensive Security & AppSec",
				Description: "Hands-on web application penetration testing, manual logic flaw hunting, and custom tool engineering.",
				Skills:      []string{"Web Pentesting", "OWASP Top 10", "SQLi / XSS Exploitation & Defense", "BOLA / IDOR Hunting", "Burp Suite Professional", "Custom Fuzzers (Go/Python)", "SAST & DAST Analysis", "Multi-Step Auth Bypass"},
			},
			{
				Icon:        "zap",
				Tag:         "Engineering Stack",
				Title:       "Backend & Systems",
				Description: "High-throughput microservices, concurrent worker pools, and automated geometric algorithms.",
				Skills:      []string{"Go (Fiber, fasthttp)", "Goroutines & Channels", "Python (Flask, OpenCV)", "C++ (Kinematics & Systems)", "Kotlin (Android Native)", "REST & Streaming APIs", "Ring-Buffered Analytics"},
			},
			{
				Icon:        "database",
				Tag:         "Data & DevOps",
				Title:       "Databases & Infra",
				Description: "High-volume relational engines, in-memory caching, containerization, and zero-downtime CD.",
				Skills:      []string{"MS SQL Server", "MySQL & SQLite", "PL/SQL Query Optimization", "Redis (Cache-Aside & Locks)", "Docker & Compose", "NGINX Reverse Proxy", "Cloudflare Tunnels", "Blue/Green Deployments"},
			},
			{
				Icon:        "radio",
				Tag:         "Embedded & Wireless",
				Title:       "Hardware & Telemetry",
				Description: "Low-power microcontroller firmware, long-range wireless protocols, and real-time sensor streams.",
				Skills:      []string{"ESP32 & ESP8266", "LoRa RF (915 MHz)", "GPS Telemetry Burst Ingestion", "IMU 6-DOF Sensor Streaming", "UART / SPI / I2C Protocols", "RF Replay Testing", "Low-Power Sleep States"},
			},
		},
		CTAEyebrow:    "Let's Connect",
		CTATitleLine1: "Ready to discuss security research,",
		CTATitleLine2: "penetration testing, or systems architecture?",
		CTALead:       "Reach out directly or download my CV to review my project and educational background.",
	}
}

// ============================================================
// 3. SECURITY TOOLS & RESEARCH CONTENT (security.html)
// ============================================================

type SecurityTool struct {
	Anchor      string   `json:"anchor"`
	Icon        string   `json:"icon"` // zap, shield, radio, database, code
	Tag         string   `json:"tag"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Chips       []string `json:"chips"`
	Status      string   `json:"status"`
	RepoURL     string   `json:"repo_url"`
}

type SecurityWriteup struct {
	Tag         string   `json:"tag"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Chips       []string `json:"chips"`
	LinkText    string   `json:"link_text"`
	LinkURL     string   `json:"link_url"`
}

type SecurityContent struct {
	MetaTitle       string            `json:"meta_title"`
	MetaDescription string            `json:"meta_description"`
	HeroEyebrow     string            `json:"hero_eyebrow"`
	HeroTitle       string            `json:"hero_title"`
	HeroLead        string            `json:"hero_lead"`
	Stats           []StatItem        `json:"stats"`
	ToolsEyebrow    string            `json:"tools_eyebrow"`
	ToolsTitle      string            `json:"tools_title"`
	ToolsLead       string            `json:"tools_lead"`
	Tools           []SecurityTool    `json:"tools"`
	WriteupsEyebrow string            `json:"writeups_eyebrow"`
	WriteupsTitle   string            `json:"writeups_title"`
	WriteupsLead    string            `json:"writeups_lead"`
	Writeups        []SecurityWriteup `json:"writeups"`
	CTAEyebrow      string            `json:"cta_eyebrow"`
	CTATitleLine1   string            `json:"cta_title_line1"`
	CTATitleLine2   string            `json:"cta_title_line2"`
	CTALead         string            `json:"cta_lead"`
	CTAPrimaryLabel string            `json:"cta_primary_label"`
	CTAPrimaryURL   string            `json:"cta_primary_url"`
}

func DefaultSecurityContent() SecurityContent {
	return SecurityContent{
		MetaTitle:       "Security Tools & Research | Jonathan Roets — AppSec Engineer",
		MetaDescription: "Open-source security tools, API fuzzers, OWASP Top 10 automated test suites, and offensive security research write-ups by Jonathan Roets.",
		HeroEyebrow:     "TRACK 01 • OFFENSIVE SECURITY",
		HeroTitle:       "Security Tools & Research.",
		HeroLead:        "Open-source offensive tooling, API vulnerability fuzzers, and deep-dive technical research on multi-step business logic bypasses and hardware RF security.",
		Stats: []StatItem{
			{Value: "OWASP", Label: "Top 10 Research"},
			{Value: "10,000+", Label: "Req/Sec Fuzzing"},
			{Value: "BOLA / IDOR", Label: "Automated Hunting"},
			{Value: "Hardware", Label: "RF Replay Auditing"},
		},
		ToolsEyebrow: "Open-Source Tooling",
		ToolsTitle:   "Offensive Tooling & Scanners.",
		ToolsLead:    "Tools built to uncover obscure attack vectors, identify race conditions, and stress-test modern web API contracts.",
		Tools: []SecurityTool{
			{
				Anchor:      "go-fuzz-api",
				Icon:        "zap",
				Tag:         "CLI Tool • High Concurrency",
				Title:       "Go-Fuzz-API • Concurrent HTTP Fuzzer",
				Description: "Engineered in Go using zero-allocation fasthttp clients and bounded worker pools. Features intelligent rate-limit detection, automated HTTP response clustering, wildcard filtering, and wordlist payload mutation. Designed to discover unlinked endpoints, hidden JSON parameters, and race conditions under intense concurrent loads.",
				Chips:       []string{"Go", "Worker Pools", "fasthttp", "Fuzzing Engine", "Rate-Limit Analyzer"},
				Status:      "Active OSS Project",
				RepoURL:     "https://github.com/JRoetscyber",
			},
			{
				Anchor:      "logic-flaw-hunter",
				Icon:        "shield",
				Tag:         "AppSec Automation • Burp Suite",
				Title:       "OWASP Top 10 Logic Flaw Hunter",
				Description: "Automated multi-session tester written in Python and Go that integrates with Burp Suite. Injects parallel token matrices across multiple user privileges to detect Broken Object-Level Authorization (BOLA/IDOR), multi-step state machine skipping (e.g. bypassing checkout payment steps), and price tampering flaws in eCommerce and financial APIs.",
				Chips:       []string{"Burp Suite API", "Python", "BOLA / IDOR", "Session Matrix", "State Tampering"},
				Status:      "Research & Tooling",
				RepoURL:     "https://github.com/JRoetscyber",
			},
			{
				Anchor:      "lora-rf-toolkit",
				Icon:        "radio",
				Tag:         "Hardware • RF Pentesting",
				Title:       "LoRa RF Telemetry Replay Toolkit",
				Description: "Firmware suite for ESP32 and SX1276 LoRa modules. Sniffs raw 915 MHz RF frames, analyzes lack of cryptographic nonce checks and replay vulnerability in agricultural and industrial telemetry collars, and provides validated cryptographic countermeasure firmware using AES-128-GCM and monotonic sequence counters.",
				Chips:       []string{"ESP32", "LoRa 915MHz", "C++ Firmware", "Packet Capture", "AES-128-GCM"},
				Status:      "Hardware Lab",
				RepoURL:     "https://github.com/JRoetscyber",
			},
			{
				Anchor:      "sqli-timing-lab",
				Icon:        "database",
				Tag:         "Database Exploitation • Research",
				Title:       "SQLi & Blind Timing Injection Lab",
				Description: "Research benchmark evaluating blind time-based SQL injection detection inside high-throughput database connection pools (MS SQL Server and PostgreSQL). Demonstrates how background query contention corrupts standard time-delay calculations and introduces an adaptive statistical threshold algorithm to extract data reliably without false positives.",
				Chips:       []string{"MS SQL Server", "PostgreSQL", "Timing Attacks", "Connection Pools", "Python"},
				Status:      "Research Paper",
				RepoURL:     "https://github.com/JRoetscyber",
			},
		},
		WriteupsEyebrow: "Technical Write-ups",
		WriteupsTitle:   "Application Security Case Studies.",
		WriteupsLead:    "In-depth dissections of architectural vulnerabilities, exploit mechanics, and remediation patterns.",
		Writeups: []SecurityWriteup{
			{
				Tag:         "Case Study • Authorization",
				Title:       "Breaking Multi-Tenant BOLA: How Inadequate Scoping Bypassed Tenant Isolation",
				Description: "An analysis of how modern SaaS multi-tenant architectures frequently enforce authentication at the gateway while relying on object IDs directly in database queries without composite tenant keys. Demonstrates how manipulating sequential or predictable UUIDv1 identifiers allowed full cross-tenant data exfiltration, and outlines the correct row-level security (RLS) and scoped query architecture to fix it.",
				Chips:       []string{"BOLA / IDOR", "Multi-Tenancy", "SQL Query Scoping", "Defense in Depth"},
				LinkText:    "Read Write-up",
				LinkURL:     "/blog",
			},
			{
				Tag:         "Case Study • Concurrency",
				Title:       "State Race Conditions: Exploiting Double-Spend in Financial Microservices",
				Description: "A deep-dive investigation into how microservice architectures with distributed state engines (e.g. coupon redemption, wallet balance debit, inventory hold) can be exploited via synchronized HTTP/2 multiplexed bursts. Explains how the time window between database read and transaction commit creates a window for double-spending, and how Redis distributed locks (Redlock) or SQL SELECT ... FOR UPDATE mitigate the threat.",
				Chips:       []string{"Race Conditions", "HTTP/2 Multiplexing", "Redis Distributed Locks", "ACID Transactions"},
				LinkText:    "Read Write-up",
				LinkURL:     "/blog",
			},
		},
		CTAEyebrow:      "AppSec Engagements",
		CTATitleLine1:   "Need a developer-led penetration test",
		CTATitleLine2:   "or application security review?",
		CTALead:         "Let's assess your APIs, review your microservice architecture, or discuss Junior Penetration Tester / AppSec opportunities.",
		CTAPrimaryLabel: "Schedule Scoped Review",
		CTAPrimaryURL:   "/contact",
	}
}

// ============================================================
// 4. SYSTEMS ENGINEERING CONTENT (engineering.html)
// ============================================================

type EngineeringBuild struct {
	Anchor          string   `json:"anchor"`
	Tag             string   `json:"tag"`
	Badge           string   `json:"badge"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Chips           []string `json:"chips"`
	HighlightsTitle string   `json:"highlights_title"`
	Highlights      []string `json:"highlights"`
}

type EngineeringContent struct {
	MetaTitle       string             `json:"meta_title"`
	MetaDescription string             `json:"meta_description"`
	HeroEyebrow     string             `json:"hero_eyebrow"`
	HeroTitle       string             `json:"hero_title"`
	HeroLead        string             `json:"hero_lead"`
	Stats           []StatItem         `json:"stats"`
	BuildsEyebrow   string             `json:"builds_eyebrow"`
	BuildsTitle     string             `json:"builds_title"`
	BuildsLead      string             `json:"builds_lead"`
	Builds          []EngineeringBuild `json:"builds"`
	CTAEyebrow      string             `json:"cta_eyebrow"`
	CTATitleLine1   string             `json:"cta_title_line1"`
	CTATitleLine2   string             `json:"cta_title_line2"`
	CTALead         string             `json:"cta_lead"`
	CTAPrimaryLabel string             `json:"cta_primary_label"`
	CTAPrimaryURL   string             `json:"cta_primary_url"`
}

func DefaultEngineeringContent() EngineeringContent {
	return EngineeringContent{
		MetaTitle:       "Systems Engineering & Builds | Jonathan Roets — Software Engineer",
		MetaDescription: "High-throughput Go Fiber microservices, enterprise ERP systems, computer vision automated quoting, and IoT telemetry engineered by Jonathan Roets.",
		HeroEyebrow:     "TRACK 02 • SYSTEMS ENGINEERING",
		HeroTitle:       "Systems & Software Builds.",
		HeroLead:        "Compiled Go microservices, enterprise prepress ERPs, geometric computer vision algorithms, and bare-metal IoT hardware engineered for raw throughput and battle-tested reliability.",
		Stats: []StatItem{
			{Value: "Sub-20ms", Label: "Server TTFB"},
			{Value: "100,000+", Label: "Req/Min Capacity"},
			{Value: "<400ms", Label: "CV Die Analysis"},
			{Value: "915 MHz", Label: "Long-Range LoRa RF"},
		},
		BuildsEyebrow: "Production Deployments",
		BuildsTitle:   "Enterprise & Embedded Systems.",
		BuildsLead:    "Real systems built to solve high-volume industrial challenges with deterministic performance.",
		Builds: []EngineeringBuild{
			{
				Anchor:          "pmd-online",
				Tag:             "Industrial ERP Platform",
				Badge:           "Production",
				Title:           "PMD Online • Packaging ERP, Die & Stereo Management System",
				Description:     "A high-throughput enterprise prepress and manufacturing management system engineered to automate the packaging lifecycle for corrugated cardboard, flexo stereos, and rotary/flatbed cutting dies. Coordinates job specs, automated proofing workflows, customer portals, and machine dispatch across multi-machine plant floors. Backed by MS SQL Server with indexed relational views, Redis singleflight caching, and compiled Go Fiber endpoints capable of handling thousands of concurrent queries with zero latency penalty.",
				Chips:           []string{"Go (Fiber)", "MS SQL Server", "PL/SQL Optimization", "Redis Cache-Aside", "Docker", "Packaging Prepress"},
				HighlightsTitle: "Architecture Highlights",
				Highlights:      []string{"Sub-15ms query execution times", "Zero-allocation JSON serialization", "Automated PDF dispatch & invoicing", "Strict RBAC & cryptographic sessions"},
			},
			{
				Anchor:          "cv-quoting",
				Tag:             "Computer Vision & Math",
				Badge:           "Automated Quoting",
				Title:           "Automated Die-Cut Layout Estimator • Computer Vision Engine",
				Description:     "Engineered in Python using OpenCV, Shapely, and ezdxf to replace hours of manual CAD measuring. Ingests raw DXF engineering drawings and high-resolution PDF vector dielines, executes morphological filtering and contour polygon decomposition, and calculates exact cutting knife lengths, crease rule footage, perforation counts, and timber board area in under 400 milliseconds. Emits instant material cost estimates directly into quoting pipelines.",
				Chips:           []string{"Python", "OpenCV", "Vector Geometry", "DXF / PDF Parsing", "Shapely Polygons", "Fast REST API"},
				HighlightsTitle: "Algorithmic Highlights",
				Highlights:      []string{"Sub-second geometric extraction", "Automated layer filtering (Cut/Crease)", "99.8% measurement accuracy vs CAD", "Eliminated 85% of manual estimator overhead"},
			},
			{
				Anchor:          "bull-agritech",
				Tag:             "Embedded Hardware & RF",
				Badge:           "IoT Telemetry",
				Title:           "B.U.L.L. Agritech • Long-Range LoRa RF Livestock Telemetry Collar",
				Description:     "Custom solar-powered collar hardware engineered for rugged South African farm environments. Uses low-power ESP32 microcontrollers paired with SX1276 LoRa transceivers to transmit GPS geofence fixes, 3-axis accelerometer kinetic metrics, and body temperature bursts over distances up to 15 kilometers without cellular dependency. Includes encrypted radio frames, deep-sleep power management (6-month battery lifespan), and an ingestion server built with Go.",
				Chips:           []string{"ESP32 Firmware", "C++ (FreeRTOS)", "LoRa 915 MHz", "GPS Telemetry", "Geofencing", "Power Optimization"},
				HighlightsTitle: "Hardware Highlights",
				Highlights:      []string{"15 km line-of-sight RF range", "Ultra-low 15µA deep sleep draw", "Real-time anti-theft geofence alerts", "Encrypted air-interface frames"},
			},
			{
				Anchor:          "golf-biomechanics",
				Tag:             "Kinematics & High-Frequency Sensors",
				Badge:           "C++ / Python",
				Title:           "High-Speed Kinetic Motion & Golf Swing Biomechanics Analyzer",
				Description:     "A real-time sports kinetic motion tracker capturing 6-DOF IMU (accelerometer and gyroscope) sensor streams at 500 Hz. Computes rotational velocity, swing plane angles, clubhead acceleration, and kinetic energy transfer curves using quaternions and complementary filtering in C++ for zero drift. Transmits data wirelessly to a Python analysis dashboard for instant bio-feedback.",
				Chips:           []string{"C++", "6-DOF IMU", "Quaternions", "Kinematics", "Real-time DSP", "Python Visualization"},
				HighlightsTitle: "Kinematic Highlights",
				Highlights:      []string{"500 Hz sensor sampling rate", "Zero-drift complementary filter", "Angular velocity & impact impulse", "Real-time kinematic curve mapping"},
			},
			{
				Anchor:          "jo4-engine",
				Tag:             "High-Concurrency Web Engine",
				Badge:           "Architecture Proof",
				Title:           "JO4 Dev Engine (This Site) • Sub-20ms Go Fiber Platform",
				Description:     "The very platform hosting this portfolio. Built with Go Fiber on fasthttp, backed by Redis cache-aside with Singleflight thundering herd protection. Analytics are ingested into a non-blocking ring buffer channel (capacity 10,000) and flushed in bulk transactions, converting 800 concurrent requests into 1 batch lock. Deployed in Docker containers with zero-downtime Blue/Green releases and routed through NGINX and Cloudflare Tunnels.",
				Chips:           []string{"Go Fiber", "Redis Singleflight", "Ring-Buffered Analytics", "Docker Blue/Green", "Cloudflare Tunnels", "Sub-20ms TTFB"},
				HighlightsTitle: "Benchmarked Specs",
				Highlights:      []string{"Zero database locks during traffic spikes", "Redis DB 1 page caching", "100/100 Core Web Vitals", "Zero-downtime container swaps"},
			},
		},
		CTAEyebrow:      "Engineering Inquiries",
		CTATitleLine1:   "Building high-throughput backends,",
		CTATitleLine2:   "IoT hardware, or custom tools?",
		CTALead:         "Let's discuss architecture, performance profiling, or software engineering roles.",
		CTAPrimaryLabel: "Get In Touch",
		CTAPrimaryURL:   "/contact",
	}
}
