package extractor

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

type EventExtractor struct {
	extractorsMap map[string]map[string]map[string][]db.Extractor
}

func NewEventExtractor() *EventExtractor {
	return &EventExtractor{
		extractorsMap: make(map[string]map[string]map[string][]db.Extractor),
	}
}

func (extractor *EventExtractor) Register(eventType, key, value string, extractors ...db.Extractor) {
	if _, ok := extractor.extractorsMap[eventType]; !ok {
		extractor.extractorsMap[eventType] = make(map[string]map[string][]db.Extractor)
	}
	if _, ok := extractor.extractorsMap[eventType][key]; !ok {
		extractor.extractorsMap[eventType][key] = make(map[string][]db.Extractor)
	}
	extractor.extractorsMap[eventType][key][value] = append(extractor.extractorsMap[eventType][key][value], extractors...)
}

func (extractor *EventExtractor) extractValue(extractors []db.Extractor, payload db.EventPayload) error {
	for _, extractor := range extractors {
		err := extractor(payload)
		if err != nil {
			return err
		}
	}
	return nil
}

func (extractor *EventExtractor) extractKey(subMap map[string][]db.Extractor, value string, payload db.EventPayload) error {
	if subMap == nil {
		return nil
	}
	for _, v := range []string{value, ""} {
		err := extractor.extractValue(subMap[v], payload)
		if err != nil {
			return err
		}
	}
	return nil
}

func (extractor *EventExtractor) extractType(subMap map[string]map[string][]db.Extractor, attributes []types.Attribute, payload db.EventPayload) error {
	if subMap == nil {
		return nil
	}
	for _, attribute := range attributes {
		for _, key := range []string{attribute.Key, ""} {
			err := extractor.extractKey(subMap[key], attribute.Value, payload)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (extractor *EventExtractor) Extract(payload db.EventPayload) error {
	events := payload.GetEvents()
	for _, event := range events {
		for _, eventType := range []string{event.Type, ""} {
			err := extractor.extractType(extractor.extractorsMap[eventType], event.Attributes, payload)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
