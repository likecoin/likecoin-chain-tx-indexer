package schema

import (
	"embed"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

//go:embed v*.sql
var fs embed.FS

func readEmbededFile(filename string) (string, error) {
	f, err := fs.Open(filename)
	if err != nil {
		logger.L.Errorw("Error when opening SQL file", "file", filename, "err", err)
		return "", err
	}
	defer f.Close()
	bz, err := io.ReadAll(f)
	if err != nil {
		logger.L.Errorw("Error when reading SQL file", "file", filename, "err", err)
		return "", err
	}
	return string(bz), nil
}

func GetVersionSQLMap() (versionSqlMap map[uint64]string, codeSchemaVersion uint64, err error) {
	files, err := fs.ReadDir(".")
	if err != nil {
		return nil, 0, err
	}
	versionSqlMap = map[uint64]string{}
	for _, fsEntry := range files {
		filename := fsEntry.Name()
		path := strings.Split(filename, "/")
		// Since the glob from go:embed is schema/v*.sql, we simply strip off the first 1 and last 4 chars
		numericPart := path[len(path)-1][1 : len(filename)-4]
		version, err := strconv.ParseUint(numericPart, 10, 64)
		if err != nil {
			logger.L.Errorw("Invalid SQL filename, expect v[schema_version].sql", "filename", filename)
			return nil, 0, err
		}
		if version == 0 {
			errMsg := "Invalid SQL schema version: 0"
			logger.L.Errorw(errMsg)
			return nil, 0, fmt.Errorf(errMsg)
		}
		if version > codeSchemaVersion {
			codeSchemaVersion = version
		}
		sql, err := readEmbededFile(filename)
		if err != nil {
			return nil, 0, err
		}
		versionSqlMap[uint64(version)] = sql
	}
	// always check existence of every version so it would be robust and won't "partially work"
	for version := uint64(1); version <= codeSchemaVersion; version++ {
		if _, ok := versionSqlMap[version]; !ok {
			logger.L.Errorw("Missing SQL file for schema version", "version", version)
			return nil, 0, fmt.Errorf("Missing SQL file for schema version %d", version)
		}
	}
	return versionSqlMap, codeSchemaVersion, nil
}
