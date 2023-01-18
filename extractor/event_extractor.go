package extractor

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

var eventExtractor = NewEventExtractor()

type EventProcessor func(payload db.EventPayload, event *types.StringEvent) error

type EventExtractor struct {
	typeKeyValueMap map[string]map[string]map[string][]EventProcessor
	typeKeyMap      map[string]map[string][]EventProcessor
	typeMap         map[string][]EventProcessor
	wildcards       []EventProcessor
}

func NewEventExtractor() *EventExtractor {
	return &EventExtractor{
		typeKeyValueMap: make(map[string]map[string]map[string][]EventProcessor),
		typeKeyMap:      make(map[string]map[string][]EventProcessor),
		typeMap:         make(map[string][]EventProcessor),
	}
}

func (e *EventExtractor) RegisterTypeKeyValue(eventType, key, value string, processor EventProcessor) {
	if _, ok := e.typeKeyValueMap[eventType]; !ok {
		e.typeKeyValueMap[eventType] = make(map[string]map[string][]EventProcessor)
	}
	if _, ok := e.typeKeyValueMap[eventType][key]; !ok {
		e.typeKeyValueMap[eventType][key] = make(map[string][]EventProcessor)
	}
	e.typeKeyValueMap[eventType][key][value] = append(e.typeKeyValueMap[eventType][key][value], processor)
}

func (e *EventExtractor) RegisterTypeKey(eventType, key string, processor EventProcessor) {
	if _, ok := e.typeKeyMap[eventType]; !ok {
		e.typeKeyMap[eventType] = make(map[string][]EventProcessor)
	}
	e.typeKeyMap[eventType][key] = append(e.typeKeyMap[eventType][key], processor)
}

func (e *EventExtractor) RegisterType(eventType string, processor EventProcessor) {
	e.typeMap[eventType] = append(e.typeMap[eventType], processor)
}

func (e *EventExtractor) RegisterAll(processor EventProcessor) {
	e.wildcards = append(e.wildcards, processor)
}

// TODO: deprecate this once all extractors are migrated to use Processor instead of db.Extractor
func (e *EventExtractor) RegisterExtractor(eventType, key, value string, extractor db.Extractor) {
	processor := func(payload db.EventPayload, event *types.StringEvent) error {
		return extractor(payload)
	}
	e.RegisterTypeKeyValue(eventType, key, value, processor)
}

func (e *EventExtractor) runProcessors(payload db.EventPayload, event *types.StringEvent, processors []EventProcessor) error {
	for _, processor := range processors {
		if err := processor(payload, event); err != nil {
			return err
		}
	}
	return nil
}

func (e *EventExtractor) extractTypeKeyValue(payload db.EventPayload, event *types.StringEvent) error {
	kvMap := e.typeKeyValueMap[event.Type]
	if kvMap == nil {
		return nil
	}
	for _, attribute := range event.Attributes {
		vMap := kvMap[attribute.Key]
		if vMap == nil {
			continue
		}
		err := e.runProcessors(payload, event, vMap[attribute.Value])
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *EventExtractor) extractTypeKey(payload db.EventPayload, event *types.StringEvent) error {
	kMap := e.typeKeyMap[event.Type]
	if kMap == nil {
		return nil
	}
	for _, attribute := range event.Attributes {
		err := e.runProcessors(payload, event, kMap[attribute.Key])
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *EventExtractor) extractType(payload db.EventPayload, event *types.StringEvent) error {
	err := e.runProcessors(payload, event, e.typeMap[event.Type])
	if err != nil {
		return err
	}
	return nil
}

func (e *EventExtractor) extractAll(payload db.EventPayload) error {
	err := e.runProcessors(payload, nil, e.wildcards)
	if err != nil {
		return err
	}
	return nil
}

func (e *EventExtractor) Extract(payload db.EventPayload) error {
	events := payload.GetEvents()
	for _, event := range events {
		type extractFuncType = func(db.EventPayload, *types.StringEvent) error
		for _, extractFunc := range []extractFuncType{e.extractTypeKeyValue, e.extractTypeKey, e.extractType} {
			err := extractFunc(payload, &event)
			if err != nil {
				return err
			}
		}
	}
	err := e.extractAll(payload)
	if err != nil {
		return err
	}
	return nil
}
