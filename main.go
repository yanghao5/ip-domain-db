package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type IPCIDR_STR struct {
	Version int64          `json:"version"`
	Rules   []IPCIDR_RULES `json:"rules"`
}

type IPCIDR_RULES struct {
	IPCIDR []string `json:"ip_cidr"`
}

func IpCidr(file_path string) []string {
	data, err := os.ReadFile(file_path)
	if err != nil {
		log.Fatal(err)
	}
	var config IPCIDR_STR

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	var ret []string
	for _, rule := range config.Rules {
		ret = append(ret, rule.IPCIDR...)
	}
	return ret
}

type DOMAIN_STR struct {
	Version int64          `json:"version"`
	Rules   []DOMAIN_RULES `json:"rules"`
}

type DOMAIN_RULES struct {
	Domain        []string `json:"domain,omitempty"`
	DomainSuffix  []string `json:"domain_suffix,omitempty"`
	DomainKeyword []string `json:"domain_keyword,omitempty"`
	DomainRegex   []string `json:"domain_regex,omitempty"`
}

type DOMAIN_ENTRY struct {
	Type  string
	Value string
}

func Domain(file_path string) []DOMAIN_ENTRY {
	data, err := os.ReadFile(file_path)
	if err != nil {
		log.Fatal(err)
	}
	var config DOMAIN_STR

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	var ret []DOMAIN_ENTRY
	for _, rule := range config.Rules {
		// 遍历 rule.Domain 中的每个元素，记录类型和值
		for _, domain := range rule.Domain {
			ret = append(ret, DOMAIN_ENTRY{
				Type:  "domain",
				Value: domain,
			})
		}

		// 遍历 rule.DomainSuffix 中的每个元素，记录类型和值
		for _, domainSuffix := range rule.DomainSuffix {
			ret = append(ret, DOMAIN_ENTRY{
				Type:  "domain_suffix",
				Value: domainSuffix,
			})
		}

		// 遍历 rule.DomainKeyword 中的每个元素，记录类型和值
		for _, domainKeyword := range rule.DomainKeyword {
			ret = append(ret, DOMAIN_ENTRY{
				Type:  "domain_keyword",
				Value: domainKeyword,
			})
		}

		for _, domainRegex := range rule.DomainRegex {
			ret = append(ret, DOMAIN_ENTRY{
				Type:  "domain_regex",
				Value: domainRegex,
			})
		}
	}
	return ret
}

func ListFiles(dirPath string) ([]string, error) {
	var filePaths []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			filePaths = append(filePaths, absPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}

// Get File Name
func GetFileName(filePath string) string {
	fileName := filepath.Base(filePath)
	extension := filepath.Ext(fileName)
	return strings.TrimSuffix(fileName, extension)
}

// write ip_cidr sqlite
// +----+----------------+
// | id | ip_cidr        |
// +----+----------------+
// | 1  | 5.62.60.5/32   |
// +----+----------------+
// | 2  | 5.62.60.6/31   |
// +----+----------------+
// | 3  | 34.99.208.0/23 |
// +----+----------------+
func WriteIpCidrToSqLite(db *sql.DB, tablename string, data []string) error {
	if db == nil {
		return errors.New("database is empty")
	}
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	// create table
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip_cidr TEXT NOT NULL
		);
	`, tablename)
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("create table error : %w", err)
	}

	// perf
	// Set PRAGMA settings outside of the transaction
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		return fmt.Errorf("set WAL mode error : %w", err)
	}
	_, err = db.Exec("PRAGMA synchronous = OFF;")
	if err != nil {
		return fmt.Errorf("set synchronous OFF error : %w", err)
	}
	_, err = db.Exec("PRAGMA temp_store = MEMORY;")
	if err != nil {
		return fmt.Errorf("set temp_store MEMORY error : %w", err)
	}

	// Enable Transactions
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("enable transactions error : %w", err)
	}
	defer tx.Rollback()

	// insert data
	insertSQL := fmt.Sprintf("INSERT INTO %s (ip_cidr) VALUES (?)", tablename)
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("prepare insert SQL error : %w", err)
	}
	defer stmt.Close()

	// insert data
	for _, value := range data {
		_, err := stmt.Exec(value)
		if err != nil {
			return fmt.Errorf("insert data error: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction error : %w", err)
	}

	return nil
}

func ProcessIpCidr(ipCidrDirPath string, db *sql.DB) {
	// meta-rules-dat/geo/geoip
	filelists, err := ListFiles(ipCidrDirPath)
	if err != nil {
		log.Fatalf("list file error : %v", err)
	}
	for _, file := range filelists {
		data := IpCidr(file)
		// write data
		fmt.Println("\033[34mProcessing: \033[32m" + file + "\033[0m")
		err = WriteIpCidrToSqLite(db, "geoip_"+GetFileName(file), data)
		if err != nil {
			log.Fatalf("write data to database error : %v", err)
		}
	}
}
func ProcessLiteIpCidr(ipCidrDirPath string, db *sql.DB) {
	// meta-rules-dat/geo/geoip
	filelists, err := ListFiles(ipCidrDirPath)
	if err != nil {
		log.Fatalf("list file error : %v", err)
	}
	for _, file := range filelists {
		data := IpCidr(file)
		// write data
		fmt.Println("\033[34mProcessing: \033[32m" + file + "\033[0m")
		err = WriteIpCidrToSqLite(db, "geoip_lite_"+GetFileName(file), data)
		if err != nil {
			log.Fatalf("write data to database error : %v", err)
		}
	}
}

// write domain
// +----+---------------+----------------+----------------+-------------------------------------------------------------------------+
// | id | domain        | domain_suffix  | domain_keyword | domain_regex                                                            |
// +----+---------------+----------------+----------------+-------------------------------------------------------------------------+
// | 1  | 1password.com |                |                |                                                                         |
// +----+---------------+----------------+----------------+-------------------------------------------------------------------------+
// | 2  |               | 4everproxy.com |                |                                                                         |
// +----+---------------+----------------+----------------+-------------------------------------------------------------------------+
// | 3  |               |                | 9to5mac        |                                                                         |
// +----+---------------+----------------+----------------+-------------------------------------------------------------------------+
// |    |               |                |                | ^github-production-release-asset-[0-9a-zA-Z]{6}\\.s3\\.amazonaws\\.com$ |
// +----+---------------+----------------+----------------+-------------------------------------------------------------------------+
func WriteDomainToSqLite(db *sql.DB, tablename string, data []DOMAIN_ENTRY) error {
	if db == nil {
		return errors.New("database is empty")
	}
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	// create table
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT,
			domain_suffix TEXT,
			domain_keyword TEXT,
			domain_regex TEXT
		);
	`, tablename)
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("create table error : %w", err)
	}

	// perf
	// Set PRAGMA settings outside of the transaction
	_, err = db.Exec(`
    PRAGMA journal_mode = WAL;
    PRAGMA synchronous = OFF;
    PRAGMA temp_store = MEMORY;
	`)
	if err != nil {
		return fmt.Errorf("set PRAGMA settings error : %w", err)
	}

	// Enable Transactions
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("enable transactions error : %w", err)
	}
	defer tx.Rollback()

	// insert data
	for _, v := range data {
		var insertSQL string
		switch v.Type {
		case "domain":
			insertSQL = fmt.Sprintf("INSERT INTO %s (domain) VALUES (?)", tablename)
		case "domain_suffix":
			insertSQL = fmt.Sprintf("INSERT INTO %s (domain_suffix) VALUES (?)", tablename)
		case "domain_keyword":
			insertSQL = fmt.Sprintf("INSERT INTO %s (domain_keyword) VALUES (?)", tablename)
		case "domain_regex":
			insertSQL = fmt.Sprintf("INSERT INTO %s (domain_regex) VALUES (?)", tablename)
		default:
			continue // Skip if type is unknown
		}
		stmt, err := tx.Prepare(insertSQL)
		if err != nil {
			return fmt.Errorf("prepare insert SQL error : %w", err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(v.Value)
		if err != nil {
			return fmt.Errorf("insert data error: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction error : %w", err)
	}

	return nil
}

func replaceString(input string) string {
	// 替换规则
	input = strings.ReplaceAll(input, "-", "_")
	input = strings.ReplaceAll(input, "@", "_")
	input = strings.ReplaceAll(input, "!", "_not_")
	input = strings.ReplaceAll(input, "__", "_")

	return input
}

func ProcessDomain(ipCidrDirPath string, db *sql.DB) {
	filelists, err := ListFiles(ipCidrDirPath)
	if err != nil {
		log.Fatalf("list file error : %v", err)
	}
	for _, file := range filelists {
		data := Domain(file)
		// write data
		fmt.Println("\033[34mProcessing: \033[32m" + file + "\033[0m")
		tablename := replaceString("domain_" + GetFileName(file))
		err = WriteDomainToSqLite(db, tablename, data)
		if err != nil {
			fmt.Println(file)
			log.Fatalf("write data to database error : %v", err)
		}
	}
}

func ProcessLiteDomain(domainDirPath string, db *sql.DB) {
	filelists, err := ListFiles(domainDirPath)
	if err != nil {
		log.Fatalf("list file error : %v", err)
	}
	for _, file := range filelists {
		data := Domain(file)
		// write data
		fmt.Println("\033[34mProcessing: \033[32m" + file + "\033[0m")
		tablename := replaceString("domain_lite_" + GetFileName(file))
		err = WriteDomainToSqLite(db, tablename, data)
		if err != nil {
			fmt.Println(file)
			log.Fatalf("write data to database error : %v", err)
		}
	}
}

// write AS sqlite
// +----+----------------+
// | id | ip_cidr        |
// +----+----------------+
// | 1  | 5.62.60.5/32   |
// +----+----------------+
// | 2  | 5.62.60.6/31   |
// +----+----------------+
// | 3  | 34.99.208.0/23 |
// +----+----------------+
func WriteASToSqLite(db *sql.DB, tablename string, data []string) error {
	if db == nil {
		return errors.New("database is empty")
	}
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	// create table
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip_cidr TEXT NOT NULL
		);
	`, tablename)
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("create table error : %w", err)
	}

	// perf
	// Set PRAGMA settings outside of the transaction
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		return fmt.Errorf("set WAL mode error : %w", err)
	}
	_, err = db.Exec("PRAGMA synchronous = OFF;")
	if err != nil {
		return fmt.Errorf("set synchronous OFF error : %w", err)
	}
	_, err = db.Exec("PRAGMA temp_store = MEMORY;")
	if err != nil {
		return fmt.Errorf("set temp_store MEMORY error : %w", err)
	}

	// Enable Transactions
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("enable transactions error : %w", err)
	}
	defer tx.Rollback()

	// insert data
	insertSQL := fmt.Sprintf("INSERT INTO %s (ip_cidr) VALUES (?)", tablename)
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("prepare insert SQL error : %w", err)
	}
	defer stmt.Close()

	// insert data
	for _, value := range data {
		_, err := stmt.Exec(value)
		if err != nil {
			return fmt.Errorf("insert data error: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction error : %w", err)
	}

	return nil
}

func ProcessAS(ipCidrDirPath string, db *sql.DB) {
	// meta-rules-dat/geo/geoip
	filelists, err := ListFiles(ipCidrDirPath)
	if err != nil {
		log.Fatalf("list file error : %v", err)
	}
	for _, file := range filelists {
		data := IpCidr(file)
		// write data
		fmt.Println("\033[34mProcessing: \033[32m" + file + "\033[0m")
		err = WriteASToSqLite(db, GetFileName(file), data)
		if err != nil {
			log.Fatalf("write data to database error : %v", err)
		}
	}
}

func main() {
	start_time := time.Now()

	// Open DB
	db, err := sql.Open("sqlite3", "./ipdomain.db")
	if err != nil {
		log.Fatalf("open db error : %v", err)
	}
	defer db.Close()

	// meta-rules-dat/geo/geoip
	//ProcessIpCidr("meta-rules-dat/geo/geoip", db)
	// meta-rules-dat/geo-lite/geoip
	//ProcessLiteIpCidr("meta-rules-dat/geo-lite/geoip", db)
	// meta-rules-dat/geo-lite/geosite
	//ProcessLiteDomain("meta-rules-dat/geo-lite/geosite", db)
	// meta-rules-dat/geo-lite/geosite
	//ProcessDomain("meta-rules-dat/geo/geosite", db)
	// meta-rules-dat/asn
	ProcessAS("meta-rules-dat/asn", db)

	duration := time.Since(start_time)
	fmt.Printf(": %v\n", duration)
	fmt.Printf("Program execution time: %v\n", duration)

}
