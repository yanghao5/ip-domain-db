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

type DOMAIN_SUFFIX_STR struct {
	Version int64                 `json:"version"`
	Rules   []DOMAIN_SUFFIX_RULES `json:"rules"`
}

type DOMAIN_SUFFIX_RULES struct {
	Domain_Suffix []string `json:"domain_suffix"`
}

func DomainSuffix(file_path string) []string {
	data, err := os.ReadFile(file_path)
	if err != nil {
		log.Fatal(err)
	}
	var config DOMAIN_SUFFIX_STR

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	var ret []string
	for _, rule := range config.Rules {
		ret = append(ret, rule.Domain_Suffix...)
	}
	return ret
}

func ListFiles(dirPath string) ([]string, error) {
	var filePaths []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // 如果出现错误，直接返回
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
func WriteIPCidrToSqLite(db *sql.DB, tablename string, data []string) error {
	if db == nil {
		return errors.New("database is empty")
	}
	if len(data) == 0 {
		return errors.New("数据列表为空")
	}

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

	// insert data
	insertSQL := fmt.Sprintf("INSERT INTO %s (ip_cidr) VALUES (?)", tablename)
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %w", err)
	}
	defer stmt.Close()

	for _, value := range data {
		_, err := stmt.Exec(value)
		if err != nil {
			return fmt.Errorf("插入数据失败: %w", err)
		}
	}
	return nil
}

// Get File Name
func GetFileName(filePath string) string {
	fileName := filepath.Base(filePath)
	extension := filepath.Ext(fileName)
	return strings.TrimSuffix(fileName, extension)
}

func main() {
	start_time := time.Now()
	fmt.Println(DomainSuffix("site.json"))
	fmt.Println(ListFiles("meta-rules-dat/geo-lite/geoip"))

	// 打开 SQLite 数据库（如果文件不存在则会自动创建）
	db, err := sql.Open("sqlite3", "./example.db")
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	// meta-rules-dat/geo/geoip
	filelists, err := ListFiles("meta-rules-dat/geo/geoip")
	if err != nil {
		log.Fatalf("list file error : %v", err)
	}
	for _, file := range filelists {
		data := IpCidr(file)
		// 写入数据到 SQLite 表
		fmt.Println("\033[34mProcessing: \033[32m" + file + "\033[0m")
		err = WriteIPCidrToSqLite(db, "geoip_"+GetFileName(file), data)
		if err != nil {
			log.Fatalf("write data to database error : %v", err)
		}
	}
	fmt.Println("Success!")
	duration := time.Since(start_time)
	fmt.Printf(": %v\n", duration)
	fmt.Printf("Program execution time: %v\n", duration)

}
