package ssh

import (
	"fmt"
	"strings"
	"sync"
)

// EnvInfo 检测到的环境信息
type EnvInfo struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Path        string   `json:"path"`              // 主路径（第一个找到的）
	Paths       []string `json:"paths,omitempty"`   // 所有找到的路径
	Status      string   `json:"status"`            // installed / not_found
}

// envRule 单个环境的检测规则
type envRule struct {
	Name        string
	Category    string
	Description string
	// 检测命令：输出路径，空表示未找到
	DetectCmd string
	// 版本命令：输出版本号
	VersionCmd string
}

// findBin 生成检测命令：依次检查 which → 常见路径
// paths 是该工具可能安装的额外路径列表
func findBin(bin string, paths ...string) string {
	cmds := []string{
		fmt.Sprintf("which %s 2>/dev/null", bin),
		fmt.Sprintf("command -v %s 2>/dev/null", bin),
	}
	for _, p := range paths {
		cmds = append(cmds, fmt.Sprintf("test -f %s && echo %s", p, p))
	}
	return strings.Join(cmds, " || ")
}

var envRules = []envRule{
	// ========== 编程语言 ==========
	{Name: "Go", Category: "编程语言", Description: "Google 开发的静态类型编译语言",
		DetectCmd:  findBin("go", "/usr/local/go/bin/go", "/snap/bin/go", "$HOME/go/bin/go"),
		VersionCmd: "go version 2>/dev/null | awk '{print $3}' | sed 's/go//'"},
	{Name: "PHP", Category: "编程语言", Description: "服务端脚本语言，广泛用于 Web 开发",
		DetectCmd:  findBin("php", "/usr/bin/php", "/usr/local/bin/php", "/usr/sbin/php", "/opt/*/bin/php"),
		VersionCmd: "php -v 2>/dev/null | head -1 | awk '{print $2}'"},
	{Name: "Python", Category: "编程语言", Description: "通用高级编程语言，广泛用于 AI/数据科学",
		DetectCmd:  findBin("python3", "python", "/usr/bin/python3", "/usr/bin/python", "/usr/local/bin/python3", "/usr/local/bin/python"),
		VersionCmd: "python3 --version 2>/dev/null | awk '{print $2}' || python --version 2>/dev/null | awk '{print $2}'"},
	{Name: "Java", Category: "编程语言", Description: "跨平台面向对象语言，企业级应用首选",
		DetectCmd:  findBin("java", "/usr/bin/java", "/usr/local/bin/java", "/usr/lib/jvm/*/bin/java", "/opt/java/*/bin/java"),
		VersionCmd: "java -version 2>&1 | head -1 | awk -F'\"' '{print $2}'"},
	{Name: "Ruby", Category: "编程语言", Description: "动态脚本语言，常用于 Web 开发和自动化",
		DetectCmd:  findBin("ruby", "/usr/bin/ruby", "/usr/local/bin/ruby", "$HOME/.rbenv/shims/ruby", "$HOME/.rvm/rubies/*/bin/ruby"),
		VersionCmd: "ruby -v 2>/dev/null | awk '{print $2}'"},
	{Name: "Rust", Category: "编程语言", Description: "系统级编程语言，注重安全与性能",
		DetectCmd:  findBin("rustc", "/usr/bin/rustc", "$HOME/.cargo/bin/rustc", "/snap/bin/rustc"),
		VersionCmd: "rustc --version 2>/dev/null | awk '{print $2}'"},
	{Name: ".NET", Category: "编程语言", Description: "微软跨平台开发框架",
		DetectCmd:  findBin("dotnet", "/usr/bin/dotnet", "/usr/local/bin/dotnet", "/usr/share/dotnet/dotnet", "/snap/bin/dotnet"),
		VersionCmd: "dotnet --version 2>/dev/null"},
	{Name: "Perl", Category: "编程语言", Description: "文本处理和系统管理脚本语言",
		DetectCmd:  findBin("perl", "/usr/bin/perl", "/usr/local/bin/perl"),
		VersionCmd: "perl -v 2>/dev/null | grep 'version=' | head -1 | sed 's/.*v\\(\\[^ ]*\\).*/\\1/'"},
	{Name: "Erlang", Category: "编程语言", Description: "函数式编程语言，高并发场景首选",
		DetectCmd:  findBin("erl", "/usr/bin/erl", "/usr/local/bin/erl", "/usr/lib/erlang/bin/erl"),
		VersionCmd: "erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().' -noshell 2>/dev/null | tr -d '\"'"},
	{Name: "Lua", Category: "编程语言", Description: "轻量级嵌入式脚本语言",
		DetectCmd:  findBin("lua", "lua5.4", "lua5.3", "/usr/bin/lua", "/usr/local/bin/lua", "/usr/bin/lua5.4", "/usr/bin/lua5.3"),
		VersionCmd: "lua -v 2>&1 | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},
	{Name: "Kotlin", Category: "编程语言", Description: "JVM 平台现代编程语言",
		DetectCmd:  findBin("kotlin", "/usr/bin/kotlin", "/usr/local/bin/kotlin", "/snap/bin/kotlin", "/opt/kotlin/bin/kotlin"),
		VersionCmd: "kotlin -version 2>&1 | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},

	// ========== 运行时 ==========
	{Name: "Node.js", Category: "运行时", Description: "基于 V8 的 JavaScript 运行时",
		DetectCmd:  findBin("node", "/usr/bin/node", "/usr/local/bin/node", "/snap/bin/node", "$HOME/.nvm/versions/node/*/bin/node", "$HOME/.volta/bin/node"),
		VersionCmd: "node -v 2>/dev/null | sed 's/^v//'"},
	{Name: "Deno", Category: "运行时", Description: "安全的 JavaScript/TypeScript 运行时",
		DetectCmd:  findBin("deno", "/usr/bin/deno", "$HOME/.deno/bin/deno", "/snap/bin/deno"),
		VersionCmd: "deno --version 2>/dev/null | head -1 | awk '{print $2}'"},
	{Name: "Bun", Category: "运行时", Description: "高性能 JavaScript 运行时与包管理器",
		DetectCmd:  findBin("bun", "/usr/bin/bun", "$HOME/.bun/bin/bun", "/snap/bin/bun"),
		VersionCmd: "bun --version 2>/dev/null"},

	// ========== 包管理器 ==========
	{Name: "npm", Category: "包管理", Description: "Node.js 包管理器",
		DetectCmd:  findBin("npm", "/usr/bin/npm", "/usr/local/bin/npm", "$HOME/.nvm/versions/node/*/bin/npm"),
		VersionCmd: "npm -v 2>/dev/null"},
	{Name: "yarn", Category: "包管理", Description: "快速可靠的 JavaScript 包管理器",
		DetectCmd:  findBin("yarn", "/usr/bin/yarn", "/usr/local/bin/yarn", "$HOME/.yarn/bin/yarn"),
		VersionCmd: "yarn -v 2>/dev/null"},
	{Name: "pnpm", Category: "包管理", Description: "高性能 Node.js 包管理器",
		DetectCmd:  findBin("pnpm", "/usr/bin/pnpm", "/usr/local/bin/pnpm", "$HOME/.local/share/pnpm/pnpm"),
		VersionCmd: "pnpm -v 2>/dev/null"},
	{Name: "pip", Category: "包管理", Description: "Python 包管理器",
		DetectCmd:  findBin("pip3", "pip", "/usr/bin/pip3", "/usr/bin/pip", "/usr/local/bin/pip3", "/usr/local/bin/pip"),
		VersionCmd: "pip3 -V 2>/dev/null | awk '{print $2}' || pip -V 2>/dev/null | awk '{print $2}'"},
	{Name: "Cargo", Category: "包管理", Description: "Rust 包管理器与构建工具",
		DetectCmd:  findBin("cargo", "/usr/bin/cargo", "$HOME/.cargo/bin/cargo"),
		VersionCmd: "cargo --version 2>/dev/null | awk '{print $2}'"},
	{Name: "Composer", Category: "包管理", Description: "PHP 依赖管理工具",
		DetectCmd:  findBin("composer", "/usr/bin/composer", "/usr/local/bin/composer", "$HOME/.composer/vendor/bin/composer"),
		VersionCmd: "composer --version 2>/dev/null | awk '{print $NF}'"},
	{Name: "gem", Category: "包管理", Description: "Ruby 包管理器",
		DetectCmd:  findBin("gem", "/usr/bin/gem", "/usr/local/bin/gem"),
		VersionCmd: "gem --version 2>/dev/null"},
	{Name: "apt", Category: "包管理", Description: "Debian/Ubuntu 软件包管理器",
		DetectCmd:  findBin("apt", "/usr/bin/apt", "/usr/bin/apt-get"),
		VersionCmd: "apt --version 2>/dev/null | head -1 | awk '{print $2}' | tr -d '()'"},
	{Name: "yum/dnf", Category: "包管理", Description: "CentOS/RHEL 软件包管理器",
		DetectCmd:  findBin("yum", "dnf", "/usr/bin/yum", "/usr/bin/dnf"),
		VersionCmd: "yum --version 2>/dev/null | head -1 || dnf --version 2>/dev/null | head -1"},
	{Name: "apk", Category: "包管理", Description: "Alpine Linux 软件包管理器",
		DetectCmd:  findBin("apk", "/sbin/apk", "/usr/sbin/apk"),
		VersionCmd: "apk --version 2>/dev/null | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},

	// ========== 数据库 ==========
	{Name: "MySQL", Category: "数据库", Description: "流行的开源关系型数据库",
		DetectCmd:  findBin("mysql", "/usr/bin/mysql", "/usr/local/bin/mysql", "/usr/sbin/mysqld", "/opt/mysql/*/bin/mysql"),
		VersionCmd: "mysql --version 2>/dev/null | awk '{print $NF}'"},
	{Name: "PostgreSQL", Category: "数据库", Description: "功能强大的开源对象关系数据库",
		DetectCmd:  findBin("psql", "/usr/bin/psql", "/usr/local/pgsql/bin/psql", "/usr/lib/postgresql/*/bin/psql"),
		VersionCmd: "psql --version 2>/dev/null | awk '{print $NF}'"},
	{Name: "MariaDB", Category: "数据库", Description: "MySQL 的开源分支",
		DetectCmd:  findBin("mariadb", "/usr/bin/mariadb", "/usr/local/bin/mariadb", "/usr/sbin/mariadbd"),
		VersionCmd: "mariadb --version 2>/dev/null | awk '{print $NF}'"},
	{Name: "Redis", Category: "数据库", Description: "高性能内存键值数据库",
		DetectCmd:  findBin("redis-server", "/usr/bin/redis-server", "/usr/local/bin/redis-server", "/usr/sbin/redis-server"),
		VersionCmd: "redis-server --version 2>/dev/null | awk '{print $3}' | cut -d= -f2"},
	{Name: "MongoDB", Category: "数据库", Description: "面向文档的 NoSQL 数据库",
		DetectCmd:  findBin("mongod", "/usr/bin/mongod", "/usr/local/bin/mongod", "/usr/bin/mongos", "/opt/mongodb/bin/mongod"),
		VersionCmd: "mongod --version 2>/dev/null | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},
	{Name: "SQLite", Category: "数据库", Description: "轻量级嵌入式关系数据库",
		DetectCmd:  findBin("sqlite3", "/usr/bin/sqlite3", "/usr/local/bin/sqlite3"),
		VersionCmd: "sqlite3 --version 2>/dev/null | awk '{print $1}'"},

	// ========== Web 服务器 ==========
	{Name: "Nginx", Category: "Web 服务器", Description: "高性能 HTTP 和反向代理服务器",
		DetectCmd:  findBin("nginx", "/usr/sbin/nginx", "/usr/local/nginx/sbin/nginx", "/usr/local/bin/nginx", "/opt/nginx/sbin/nginx"),
		VersionCmd: "nginx -v 2>&1 | awk -F/ '{print $2}'"},
	{Name: "Apache", Category: "Web 服务器", Description: "世界使用最广泛的 Web 服务器",
		DetectCmd:  findBin("httpd", "apache2", "/usr/sbin/httpd", "/usr/sbin/apache2", "/usr/local/apache2/bin/httpd"),
		VersionCmd: "httpd -v 2>/dev/null | sed 's/.*Apache\\/\\([0-9.]*\\).*/\\1/' || apache2 -v 2>/dev/null | sed 's/.*Apache\\/\\([0-9.]*\\).*/\\1/'"},
	{Name: "Caddy", Category: "Web 服务器", Description: "自动 HTTPS 的现代 Web 服务器",
		DetectCmd:  findBin("caddy", "/usr/bin/caddy", "/usr/local/bin/caddy", "/snap/bin/caddy"),
		VersionCmd: "caddy version 2>/dev/null | awk '{print $1}' | sed 's/^v//'"},
	{Name: "Tomcat", Category: "Web 服务器", Description: "Java Servlet 容器",
		DetectCmd:  findBin("catalina.sh", "/usr/bin/catalina.sh", "/opt/tomcat/bin/catalina.sh", "/opt/*/bin/catalina.sh", "/usr/share/tomcat*/bin/catalina.sh"),
		VersionCmd: "catalina.sh version 2>/dev/null | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},

	// ========== 容器 & 编排 ==========
	{Name: "Docker", Category: "容器", Description: "容器化应用部署平台",
		DetectCmd:  findBin("docker", "/usr/bin/docker", "/usr/local/bin/docker", "/snap/bin/docker"),
		VersionCmd: "docker --version 2>/dev/null | awk '{for(i=3;i<=NF;i++){gsub(/,/,\"\",$i); if($i ~ /^[0-9]/){print $i; exit}}}'"},
	{Name: "Podman", Category: "容器", Description: "无守护进程的容器引擎",
		DetectCmd:  findBin("podman", "/usr/bin/podman", "/usr/local/bin/podman"),
		VersionCmd: "podman --version 2>/dev/null | awk '{print $2}'"},
	{Name: "containerd", Category: "容器", Description: "工业级容器运行时",
		DetectCmd:  findBin("containerd", "/usr/bin/containerd", "/usr/local/bin/containerd", "/usr/sbin/containerd"),
		VersionCmd: "containerd --version 2>/dev/null | awk '{print $3}'"},
	{Name: "kubectl", Category: "容器", Description: "Kubernetes 命令行工具",
		DetectCmd:  findBin("kubectl", "/usr/bin/kubectl", "/usr/local/bin/kubectl", "/snap/bin/kubectl", "$HOME/.local/bin/kubectl"),
		VersionCmd: "kubectl version --client 2>/dev/null | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^v?[0-9]/){gsub(/^v/,\"\",$i); print $i; exit}}'"},
	{Name: "Helm", Category: "容器", Description: "Kubernetes 包管理器",
		DetectCmd:  findBin("helm", "/usr/bin/helm", "/usr/local/bin/helm", "/snap/bin/helm"),
		VersionCmd: "helm version --short 2>/dev/null | sed 's/^v//'"},

	// ========== DevOps ==========
	{Name: "Terraform", Category: "DevOps", Description: "基础设施即代码工具",
		DetectCmd:  findBin("terraform", "/usr/bin/terraform", "/usr/local/bin/terraform", "/snap/bin/terraform"),
		VersionCmd: "terraform version 2>/dev/null | head -1 | awk '{print $2}' | sed 's/^v//'"},
	{Name: "Ansible", Category: "DevOps", Description: "自动化运维配置管理工具",
		DetectCmd:  findBin("ansible", "/usr/bin/ansible", "/usr/local/bin/ansible", "$HOME/.local/bin/ansible"),
		VersionCmd: "ansible --version 2>/dev/null | head -1 | awk '{print $2}'"},

	// ========== 构建工具 ==========
	{Name: "Make", Category: "构建工具", Description: "经典构建自动化工具",
		DetectCmd:  findBin("make", "/usr/bin/make", "/usr/local/bin/make"),
		VersionCmd: "make --version 2>/dev/null | head -1 | awk '{print $3}'"},
	{Name: "CMake", Category: "构建工具", Description: "跨平台构建系统",
		DetectCmd:  findBin("cmake", "/usr/bin/cmake", "/usr/local/bin/cmake", "/snap/bin/cmake"),
		VersionCmd: "cmake --version 2>/dev/null | head -1 | awk '{print $3}'"},
	{Name: "GCC", Category: "构建工具", Description: "GNU 编译器套件",
		DetectCmd:  findBin("gcc", "/usr/bin/gcc", "/usr/local/bin/gcc"),
		VersionCmd: "gcc --version 2>/dev/null | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},
	{Name: "G++", Category: "构建工具", Description: "GNU C++ 编译器",
		DetectCmd:  findBin("g++", "/usr/bin/g++", "/usr/local/bin/g++"),
		VersionCmd: "g++ --version 2>/dev/null | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},
	{Name: "Clang", Category: "构建工具", Description: "LLVM C/C++ 编译器",
		DetectCmd:  findBin("clang", "/usr/bin/clang", "/usr/local/bin/clang"),
		VersionCmd: "clang --version 2>/dev/null | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},
	{Name: "Gradle", Category: "构建工具", Description: "基于 Groovy/Kotlin 的构建工具",
		DetectCmd:  findBin("gradle", "/usr/bin/gradle", "/usr/local/bin/gradle", "/opt/gradle/bin/gradle", "$HOME/.sdkman/candidates/gradle/*/bin/gradle"),
		VersionCmd: "gradle --version 2>/dev/null | grep 'Gradle' | awk '{print $2}'"},
	{Name: "Maven", Category: "构建工具", Description: "Java 项目管理和构建工具",
		DetectCmd:  findBin("mvn", "/usr/bin/mvn", "/usr/local/bin/mvn", "/opt/maven/bin/mvn", "$HOME/.sdkman/candidates/maven/*/bin/mvn"),
		VersionCmd: "mvn --version 2>/dev/null | head -1 | awk '{print $3}'"},

	// ========== 版本控制 ==========
	{Name: "Git", Category: "版本控制", Description: "分布式版本控制系统",
		DetectCmd:  findBin("git", "/usr/bin/git", "/usr/local/bin/git", "/snap/bin/git"),
		VersionCmd: "git --version 2>/dev/null | awk '{print $3}'"},
	{Name: "SVN", Category: "版本控制", Description: "集中式版本控制系统",
		DetectCmd:  findBin("svn", "/usr/bin/svn", "/usr/local/bin/svn"),
		VersionCmd: "svn --version --quiet 2>/dev/null"},

	// ========== 网络工具 ==========
	{Name: "curl", Category: "网络工具", Description: "命令行数据传输工具",
		DetectCmd:  findBin("curl", "/usr/bin/curl", "/usr/local/bin/curl", "/snap/bin/curl"),
		VersionCmd: "curl --version 2>/dev/null | head -1 | awk '{print $2}'"},
	{Name: "wget", Category: "网络工具", Description: "网络下载工具",
		DetectCmd:  findBin("wget", "/usr/bin/wget", "/usr/local/bin/wget"),
		VersionCmd: "wget --version 2>/dev/null | head -1 | awk '{print $3}'"},
	{Name: "OpenSSH", Category: "网络工具", Description: "安全远程连接协议套件",
		DetectCmd:  findBin("ssh", "/usr/bin/ssh", "/usr/local/bin/ssh"),
		VersionCmd: "ssh -V 2>&1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},
	{Name: "rsync", Category: "网络工具", Description: "快速增量文件同步工具",
		DetectCmd:  findBin("rsync", "/usr/bin/rsync", "/usr/local/bin/rsync"),
		VersionCmd: "rsync --version 2>/dev/null | head -1 | awk '{print $3}'"},

	// ========== 安全工具 ==========
	{Name: "OpenSSL", Category: "安全工具", Description: "TLS/SSL 加密库与工具",
		DetectCmd:  findBin("openssl", "/usr/bin/openssl", "/usr/local/bin/openssl", "/usr/sbin/openssl"),
		VersionCmd: "openssl version 2>/dev/null | awk '{print $2}'"},
	{Name: "GPG", Category: "安全工具", Description: "GNU 隐私保护加密工具",
		DetectCmd:  findBin("gpg", "gpg2", "/usr/bin/gpg", "/usr/bin/gpg2", "/usr/local/bin/gpg"),
		VersionCmd: "gpg --version 2>/dev/null | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},
	{Name: "Fail2ban", Category: "安全工具", Description: "入侵防御框架",
		DetectCmd:  findBin("fail2ban-client", "/usr/bin/fail2ban-client", "/usr/local/bin/fail2ban-client"),
		VersionCmd: "fail2ban-client --version 2>/dev/null | head -1 | awk '{print $2}'"},

	// ========== 系统工具 ==========
	{Name: "tmux", Category: "系统工具", Description: "终端复用器",
		DetectCmd:  findBin("tmux", "/usr/bin/tmux", "/usr/local/bin/tmux", "/snap/bin/tmux"),
		VersionCmd: "tmux -V 2>/dev/null | awk '{print $2}'"},
	{Name: "screen", Category: "系统工具", Description: "终端会话管理器",
		DetectCmd:  findBin("screen", "/usr/bin/screen", "/usr/local/bin/screen"),
		VersionCmd: "screen -v 2>/dev/null | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^[0-9]/){print $i; exit}}'"},
	{Name: "htop", Category: "系统工具", Description: "交互式进程查看器",
		DetectCmd:  findBin("htop", "/usr/bin/htop", "/usr/local/bin/htop", "/snap/bin/htop"),
		VersionCmd: "htop --version 2>/dev/null | head -1 | awk '{print $2}'"},
	{Name: "jq", Category: "系统工具", Description: "轻量级命令行 JSON 处理器",
		DetectCmd:  findBin("jq", "/usr/bin/jq", "/usr/local/bin/jq", "/snap/bin/jq"),
		VersionCmd: "jq --version 2>/dev/null | sed 's/^jq-//'"},
	{Name: "Vim", Category: "系统工具", Description: "高度可配置的文本编辑器",
		DetectCmd:  findBin("vim", "vi", "/usr/bin/vim", "/usr/bin/vi", "/usr/local/bin/vim"),
		VersionCmd: "vim --version 2>/dev/null | head -1 | awk '{for(i=2;i<=NF;i++) if($i ~ /^V?[0-9]/){gsub(/^V/,\"\",$i); print $i; exit}}'"},
	{Name: "Neovim", Category: "系统工具", Description: "现代化 Vim 文本编辑器",
		DetectCmd:  findBin("nvim", "/usr/bin/nvim", "/usr/local/bin/nvim", "/snap/bin/nvim"),
		VersionCmd: "nvim --version 2>/dev/null | head -1 | awk '{print $2}'"},
}

// GetSystemEnvironments 并发检测远程服务器上的所有环境
func (c *SSHClient) GetSystemEnvironments() ([]EnvInfo, error) {
	if c.isConnected.Load() == 0 {
		return nil, fmt.Errorf("SSH 未连接")
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	results := make([]EnvInfo, 0, len(envRules))

	sem := make(chan struct{}, 12)

	for _, rule := range envRules {
		wg.Add(1)
		go func(r envRule) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			env := c.detectOne(r)
			mu.Lock()
			results = append(results, env)
			mu.Unlock()
		}(rule)
	}

	wg.Wait()

	var installed, notFound []EnvInfo
	for _, e := range results {
		if e.Status == "installed" {
			installed = append(installed, e)
		} else {
			notFound = append(notFound, e)
		}
	}

	return append(installed, notFound...), nil
}

// detectOne 检测单个环境
func (c *SSHClient) detectOne(r envRule) EnvInfo {
	env := EnvInfo{
		Name:        r.Name,
		Category:    r.Category,
		Description: r.Description,
		Status:      "not_found",
	}

	// 检测路径（DetectCmd 已包含 which + 常见路径）
	pathResult, err := c.ExecuteCommand(r.DetectCmd)
	if err == nil && pathResult.Success {
		// 收集所有找到的路径（去重）
		paths := uniqueLines(pathResult.Stdout)
		if len(paths) > 0 {
			env.Path = paths[0]
			if len(paths) > 1 {
				env.Paths = paths
			}
			env.Status = "installed"
		}
	}

	// 如果 DetectCmd 失败，尝试包管理器查询
	if env.Status == "not_found" {
		binName := extractBinaryName(r.DetectCmd)
		pkgCmd := fmt.Sprintf(
			`(dpkg -L %s 2>/dev/null | grep '/bin/%s$' | head -1) || (rpm -ql %s 2>/dev/null | grep '/bin/%s$' | head -1)`,
			r.Name, binName, r.Name, binName,
		)
		pkgResult, err := c.ExecuteCommand(pkgCmd)
		if err == nil && pkgResult.Success {
			path := strings.TrimSpace(pkgResult.Stdout)
			if path != "" {
				env.Path = path
				env.Status = "installed"
			}
		}
		// 如果上面没找到具体路径，查询版本信息
		if env.Status == "not_found" {
			verCmd := fmt.Sprintf(
				`(dpkg -s %s 2>/dev/null | grep '^Version:' | awk '{print $2}') || (rpm -q %s 2>/dev/null | sed 's/^.*-//')`,
				r.Name, r.Name,
			)
			verResult, err := c.ExecuteCommand(verCmd)
			if err == nil && verResult.Success {
				ver := strings.TrimSpace(verResult.Stdout)
				if ver != "" {
					env.Version = ver
					env.Path = "(via " + r.Name + " package)"
					env.Status = "installed"
				}
			}
		}
	}

	// 获取版本
	if env.Status == "installed" && r.VersionCmd != "" && env.Version == "" {
		verResult, err := c.ExecuteCommand(r.VersionCmd)
		if err == nil && verResult.Success {
			env.Version = strings.TrimSpace(verResult.Stdout)
		}
	}

	return env
}

// uniqueLines 从多行输出中提取所有非空、去重的行
func uniqueLines(output string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	return result
}

// extractBinaryName 从 DetectCmd 中提取要检测的二进制名称
func extractBinaryName(cmd string) string {
	fields := strings.Fields(cmd)
	for i, f := range fields {
		if f == "which" || f == "command" || f == "type" {
			for j := i + 1; j < len(fields); j++ {
				next := fields[j]
				if next == "-v" || next == "-p" || next == "-n" || next == "2>/dev/null" || next == "||" || next == "&&" {
					continue
				}
				if strings.HasPrefix(next, "-") {
					continue
				}
				return next
			}
		}
	}
	for _, f := range fields {
		if !strings.HasPrefix(f, "-") && f != "2>/dev/null" && f != "||" && f != "&&" && f != "|" {
			return f
		}
	}
	return ""
}
