package ssh

import (
	"database/sql"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// DBConnection 数据库连接（通过 SSH 隧道）
type DBConnection struct {
	db       *sql.DB
	listener net.Listener
	sshConn  *SSHClient
	dbType   string
	dbHost   string
	dbPort   int
}

// DatabaseService 数据库服务（Go 原生驱动 + SSH 隧道）
type DatabaseService struct {
	mu      sync.RWMutex
	sshSvc  *SSHService
	conns   map[string]*DBConnection // key = connID:dbType:host:port
}

// NewDatabaseService 创建数据库服务
func NewDatabaseService(sshSvc *SSHService) *DatabaseService {
	return &DatabaseService{
		sshSvc: sshSvc,
		conns:  make(map[string]*DBConnection),
	}
}

// connKey 生成连接缓存 key（包含数据库名，不同数据库用不同连接）
func connKey(connID, dbType, host string, port int, database string) string {
	return fmt.Sprintf("%s:%s:%s:%d:%s", connID, dbType, host, port, database)
}

// openDB 通过 SSH 隧道打开数据库连接
func (s *DatabaseService) openDB(connID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase string) (*sql.DB, error) {
	key := connKey(connID, dbType, dbHost, dbPort, dbDatabase)

	// 使用写锁防止并发创建同一连接
	s.mu.Lock()

	// 检查缓存
	if existing, ok := s.conns[key]; ok && existing.db != nil {
		db := existing.db
		s.mu.Unlock()
		// 缓存命中，直接返回（Ping 在锁外执行避免死锁）
		if err := db.Ping(); err == nil {
			return db, nil
		}
		// 连接已失效，重新创建
		s.mu.Lock()
		if old, ok := s.conns[key]; ok {
			old.db.Close()
			if old.listener != nil {
				old.listener.Close()
			}
			delete(s.conns, key)
		}
	}

	// 获取 SSH 客户端
	client, err := s.sshSvc.GetClient(connID)
	if err != nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("SSH 连接不存在: %v", err)
	}
	if !client.IsConnected() {
		s.mu.Unlock()
		return nil, fmt.Errorf("SSH 连接已断开")
	}

	// 创建 SSH 隧道
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("创建本地监听失败: %v", err)
	}
	localPort := listener.Addr().(*net.TCPAddr).Port

	// 启动隧道转发
	go s.forwardTunnel(listener, client, dbHost, dbPort)

	// 等待隧道就绪
	time.Sleep(100 * time.Millisecond)

	// 构建 DSN
	var dsn string
	switch dbType {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(127.0.0.1:%d)/%s?timeout=10s&parseTime=true",
			dbUser, dbPassword, localPort, dbDatabase)
	case "postgresql":
		dsn = fmt.Sprintf("host=127.0.0.1 port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=10",
			localPort, dbUser, dbPassword, dbDatabase)
	default:
		listener.Close()
		s.mu.Unlock()
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	// 打开数据库连接
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		listener.Close()
		s.mu.Unlock()
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		listener.Close()
		s.mu.Unlock()
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 缓存连接
	s.conns[key] = &DBConnection{
		db:       db,
		listener: listener,
		sshConn:  client,
		dbType:   dbType,
		dbHost:   dbHost,
		dbPort:   dbPort,
	}
	s.mu.Unlock()

	return db, nil
}

// forwardTunnel SSH 隧道转发
func (s *DatabaseService) forwardTunnel(listener net.Listener, client *SSHClient, remoteHost string, remotePort int) {
	remoteAddr := fmt.Sprintf("%s:%d", remoteHost, remotePort)

	for {
		localConn, err := listener.Accept()
		if err != nil {
			return // listener 已关闭
		}

		remoteConn, err := client.client.Dial("tcp", remoteAddr)
		if err != nil {
			localConn.Close()
			continue
		}

		go func() {
			defer localConn.Close()
			defer remoteConn.Close()
			done := make(chan struct{})
			go func() {
				buf := make([]byte, 32*1024)
				io.CopyBuffer(remoteConn, localConn, buf)
				close(done)
			}()
			buf := make([]byte, 32*1024)
			io.CopyBuffer(localConn, remoteConn, buf)
			<-done
		}()
	}
}

// CloseDB 关闭指定数据库连接
func (s *DatabaseService) CloseDB(connID, dbType, dbHost string, dbPort int, dbDatabase string) {
	key := connKey(connID, dbType, dbHost, dbPort, dbDatabase)
	s.mu.Lock()
	if existing, ok := s.conns[key]; ok {
		if existing.db != nil {
			existing.db.Close()
		}
		if existing.listener != nil {
			existing.listener.Close()
		}
		delete(s.conns, key)
	}
	s.mu.Unlock()
}

// ========== 公共 API ==========

// TestConnection 测试数据库连接
func (s *DatabaseService) TestConnection(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword string) error {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, "")
	if err != nil {
		return err
	}
	return db.Ping()
}

// ListDatabases 列出数据库
func (s *DatabaseService) ListDatabases(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword string) ([]string, error) {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, "")
	if err != nil {
		return nil, err
	}

	var query string
	switch dbType {
	case "mysql":
		query = "SHOW DATABASES"
	case "postgresql":
		query = "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname"
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		databases = append(databases, name)
	}
	return databases, nil
}

// ListTables 列出表
func (s *DatabaseService) ListTables(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase string) ([]string, error) {
	log.Info("[DB] ListTables", "sshConnID", sshConnID, "dbType", dbType, "dbHost", dbHost, "dbPort", dbPort, "dbUser", dbUser, "dbDatabase", dbDatabase)

	if dbDatabase == "" {
		return nil, fmt.Errorf("未选择数据库")
	}

	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		return nil, err
	}

	var query string
	switch dbType {
	case "mysql":
		query = "SHOW TABLES"
	case "postgresql":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename"
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		tables = append(tables, name)
	}
	return tables, nil
}

// GetTableStructure 获取表结构
func (s *DatabaseService) GetTableStructure(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase, tableName string) ([][]string, []string, error) {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		return nil, nil, err
	}

	var query string
	switch dbType {
	case "mysql":
		query = fmt.Sprintf("DESCRIBE `%s`", tableName)
	case "postgresql":
		query = fmt.Sprintf(`
			SELECT column_name, data_type, is_nullable, column_default, character_maximum_length
			FROM information_schema.columns
			WHERE table_name = '%s' AND table_schema = 'public'
			ORDER BY ordinal_position`, tableName)
	default:
		return nil, nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	return s.queryToRows(db, query)
}

// GetTableData 浏览表数据（分页）
func (s *DatabaseService) GetTableData(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase, tableName string, page, pageSize int) ([][]string, []string, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		log.Error("[DB] GetTableData openDB 失败", "error", err)
		return nil, nil, err
	}

	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", tableName, pageSize, offset)
	log.Info("[DB] GetTableData", "query", query)

	rows, columns, err := s.queryToRows(db, query)
	if err != nil {
		log.Error("[DB] GetTableData 查询失败", "error", err, "query", query)
		return nil, nil, err
	}

	log.Info("[DB] GetTableData 结果", "rows", len(rows), "columns", len(columns), "query", query)
	return rows, columns, nil
}

// GetTableCount 获取表记录数
func (s *DatabaseService) GetTableCount(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase, tableName string) (int64, error) {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		return 0, err
	}

	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	err = db.QueryRow(query).Scan(&count)
	return count, err
}

// GetTableIndexes 获取表索引
func (s *DatabaseService) GetTableIndexes(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase, tableName string) ([][]string, []string, error) {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		return nil, nil, err
	}

	var query string
	switch dbType {
	case "mysql":
		query = fmt.Sprintf("SHOW INDEX FROM `%s`", tableName)
	case "postgresql":
		query = fmt.Sprintf(`
			SELECT indexname, indexdef
			FROM pg_indexes
			WHERE tablename = '%s' AND schemaname = 'public'`, tableName)
	default:
		return nil, nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	return s.queryToRows(db, query)
}

// ExecuteQuery 执行查询
func (s *DatabaseService) ExecuteQuery(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase, query string) ([][]string, []string, error) {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		return nil, nil, err
	}

	return s.queryToRows(db, query)
}

// GetDatabaseSize 获取数据库大小
func (s *DatabaseService) GetDatabaseSize(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase string) (string, error) {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		return "", err
	}

	var query string
	switch dbType {
	case "mysql":
		query = "SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) FROM information_schema.tables WHERE table_schema = ?"
	case "postgresql":
		query = "SELECT pg_size_pretty(pg_database_size($1))"
	default:
		return "", fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	var size string
	err = db.QueryRow(query, dbDatabase).Scan(&size)
	if err != nil {
		return "", err
	}
	return size, nil
}

// GetTableStatus 获取表状态
func (s *DatabaseService) GetTableStatus(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase string) ([][]string, []string, error) {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		return nil, nil, err
	}

	var query string
	switch dbType {
	case "mysql":
		query = fmt.Sprintf("SHOW TABLE STATUS FROM `%s`", dbDatabase)
	case "postgresql":
		query = fmt.Sprintf(`
			SELECT relname AS table_name, n_live_tup AS row_count, pg_size_pretty(pg_total_relation_size(relid)) AS total_size
			FROM pg_stat_user_tables
			WHERE schemaname = '%s'
			ORDER BY n_live_tup DESC`, dbDatabase)
	default:
		return nil, nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	return s.queryToRows(db, query)
}

// TruncateTable 清空表
func (s *DatabaseService) TruncateTable(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase, tableName string) error {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", tableName))
	return err
}

// DropTable 删除表
func (s *DatabaseService) DropTable(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase, tableName string) error {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, dbDatabase)
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("DROP TABLE `%s`", tableName))
	return err
}

// CreateDatabase 创建数据库
func (s *DatabaseService) CreateDatabase(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase string) error {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, "")
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE `%s`", dbDatabase))
	return err
}

// DropDatabase 删除数据库
func (s *DatabaseService) DropDatabase(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, dbDatabase string) error {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, "")
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE `%s`", dbDatabase))
	return err
}

// DetectDatabases 自动检测服务器上可用的数据库
func (s *DatabaseService) DetectDatabases(sshConnID string) []map[string]string {
	client, err := s.sshSvc.GetClient(sshConnID)
	if err != nil || !client.IsConnected() {
		return nil
	}

	var results []map[string]string

	// 检测 MySQL
	if out, err := client.ExecuteCommand("mysql --version 2>&1 || which mysql 2>&1"); err == nil && out != nil {
		version := strings.TrimSpace(out.Stdout)
		if version != "" && !strings.Contains(version, "not found") {
			results = append(results, map[string]string{
				"type": "mysql", "name": "MySQL", "port": "3306", "version": version,
			})
		}
	}

	// 检测 PostgreSQL
	if out, err := client.ExecuteCommand("psql --version 2>&1 || which psql 2>&1"); err == nil && out != nil {
		version := strings.TrimSpace(out.Stdout)
		if version != "" && !strings.Contains(version, "not found") {
			results = append(results, map[string]string{
				"type": "postgresql", "name": "PostgreSQL", "port": "5432", "version": version,
			})
		}
	}

	// 检测 Redis
	if out, err := client.ExecuteCommand("redis-cli --version 2>&1 || which redis-cli 2>&1"); err == nil && out != nil {
		version := strings.TrimSpace(out.Stdout)
		if version != "" && !strings.Contains(version, "not found") {
			results = append(results, map[string]string{
				"type": "redis", "name": "Redis", "port": "6379", "version": version,
			})
		}
	}

	// 检测正在运行的服务端口
	portMap := map[string]string{
		"3306": "mysql", "5432": "postgresql", "6379": "redis", "27017": "mongodb",
	}
	out, _ := client.ExecuteCommand("ss -tlnp 2>/dev/null | grep -oE ':(3306|5432|6379|27017)' || netstat -tlnp 2>/dev/null | grep -oE ':(3306|5432|6379|27017)' || true")
	for _, line := range strings.Split(out.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		port := strings.TrimPrefix(line, ":")
		if dbType, ok := portMap[port]; ok {
			for i, r := range results {
				if r["type"] == dbType {
					results[i]["status"] = "running"
				}
			}
		}
	}

	return results
}

// queryToRows 执行查询并返回结构化结果
func (s *DatabaseService) queryToRows(db *sql.DB, query string) ([][]string, []string, error) {
	rows, err := db.Query(query)
	if err != nil {
		log.Error("[DB] queryToRows 查询失败", "error", err, "query", query)
		return nil, nil, fmt.Errorf("查询失败: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Error("[DB] queryToRows 获取列名失败", "error", err)
		return nil, nil, err
	}

	log.Info("[DB] queryToRows 列名", "columns", columns)

	var result [][]string
	rowCount := 0
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Error("[DB] queryToRows Scan 失败", "error", err, "row", rowCount)
			continue
		}

		row := make([]string, len(columns))
		for i, v := range values {
			if v == nil {
				row[i] = "NULL"
			} else {
				switch val := v.(type) {
				case []byte:
					row[i] = string(val)
				case time.Time:
					row[i] = val.Format("2006-01-02 15:04:05")
				default:
					row[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		result = append(result, row)
		rowCount++
	}

	log.Info("[DB] queryToRows 完成", "rows", rowCount, "query", query)
	return result, columns, nil
}

// GetProcessList 获取数据库进程列表
func (s *DatabaseService) GetProcessList(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword string) ([][]string, []string, error) {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, "")
	if err != nil {
		return nil, nil, err
	}

	var query string
	switch dbType {
	case "mysql":
		query = "SHOW PROCESSLIST"
	case "postgresql":
		query = "SELECT pid, usename, datname, state, query, query_start FROM pg_stat_activity ORDER BY query_start DESC LIMIT 50"
	default:
		return nil, nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	return s.queryToRows(db, query)
}

// KillQuery 终止查询
func (s *DatabaseService) KillQuery(sshConnID, dbType, dbHost string, dbPort int, dbUser, dbPassword, queryID string) error {
	db, err := s.openDB(sshConnID, dbType, dbHost, dbPort, dbUser, dbPassword, "")
	if err != nil {
		return err
	}

	var query string
	switch dbType {
	case "mysql":
		query = fmt.Sprintf("KILL %s", queryID)
	case "postgresql":
		query = fmt.Sprintf("SELECT pg_terminate_backend(%s)", queryID)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	_, err = db.Exec(query)
	return err
}
