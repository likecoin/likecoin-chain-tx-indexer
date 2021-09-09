package util

import (
	"fmt"
	"hash/crc64"
)

var partitionTable = crc64.MakeTable(crc64.ISO)

func GetEventHashes(events []Event) []int64 {
	eventHashes := []int64{}
	for _, event := range events {
		for _, attr := range event.Attributes {
			s := fmt.Sprintf("%s.%s=\"%s\"", event.Type, attr.Key, attr.Value)
			hash := int64(crc64.Checksum([]byte(s), partitionTable))
			eventHashes = append(eventHashes, hash)
		}
	}
	return eventHashes
}
