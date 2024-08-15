package rest

import (
	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

type copyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (cw copyWriter) Write(b []byte) (int, error) {
	return cw.body.Write(b)
}

func (cw *copyWriter) WriteHeader(statusCode int) {
	// Remove content-length header written by proxy
	cw.Header().Del("Content-length")
	cw.Header().Set("Transfer-Encoding", "chunked")
	cw.ResponseWriter.WriteHeader(statusCode)
}

func filterContentFingerprints() gin.HandlerFunc {
	return func(c *gin.Context) {
		isBrowser := false
		switch c.NegotiateFormat(gin.MIMEJSON, gin.MIMEHTML) {
		case gin.MIMEHTML:
			isBrowser = true
		case gin.MIMEJSON:
			isBrowser = false
		}

		if !isBrowser {
			c.Next()
			return
		}

		wb := &copyWriter{
			body:           &bytes.Buffer{},
			ResponseWriter: c.Writer,
		}
		c.Writer = wb
		c.Next()

		originBody := wb.body
		originBodyBytes := originBody.Bytes()

		var jsonObject map[string]any
		err := json.Unmarshal(originBodyBytes, &jsonObject)
		if err != nil {
			_, e := wb.ResponseWriter.Write(originBodyBytes)
			if e != nil {
				logger.L.Error(e)
			}
			return
		}

		records, ok := jsonObject["records"].([]interface{})
		if !ok {
			_, e := wb.ResponseWriter.Write(originBodyBytes)
			if e != nil {
				logger.L.Error(e)
			}
			return
		}

		for index, record := range records {
			obj, ok := record.(map[string]any)
			if !ok {
				continue
			}
			data, ok := obj["data"].(map[string]any)
			if !ok {
				continue
			}
			if data["contentFingerprints"] != nil {
				delete(data, "contentFingerprints")
			}
			if data["contentMetadata"] != nil {
				contentMetadata, ok := data["contentMetadata"].(map[string]any)
				if ok {
					if contentMetadata["sameAs"] != nil {
						delete(contentMetadata, "sameAs")
					}
					data["contentMetadata"] = contentMetadata
				}
			}
			obj["data"] = data
			records[index] = obj
		}
		jsonObject["records"] = records

		newBody, err := json.Marshal(jsonObject)
		if err != nil {
			_, e := wb.ResponseWriter.Write(originBodyBytes)
			if e != nil {
				logger.L.Error(e)
			}
			return
		}
		_, e := wb.ResponseWriter.Write(newBody)
		if e != nil {
			logger.L.Error(e)
		}
	}
}
