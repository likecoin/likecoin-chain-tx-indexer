package extractor

import (
	"encoding/json"
	"strconv"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

var eventExtractor = NewEventExtractor()

func extractAuthzEventsList(events types.StringEvents) (db.EventsList, error) {
	authzEvents := make(map[int]types.StringEvents)
	specialEvents := []types.StringEvent{}
	maxIndex := 0
	for _, event := range events {
		authzAttrsMap := make(map[int][]types.Attribute)
		accumulatedAttrs := []types.Attribute{}
		for _, attr := range event.Attributes {
			if event.Type == "message" &&
				attr.Key == "action" &&
				attr.Value == "/cosmos.authz.v1beta1.MsgExec" {
				// skip this so authz messages won't be recursively extracted
				// TODO: do we need recursive extraction?
				continue
			}
			if attr.Key != "authz_msg_index" {
				accumulatedAttrs = append(accumulatedAttrs, attr)
				continue
			}
			// authz_msg_index is indicating the msg index of previous attributes (which are in accumulatedAttrs)
			msgIndexUint64, err := strconv.ParseUint(attr.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			msgIndex := int(msgIndexUint64)
			authzAttrsMap[msgIndex] = append(authzAttrsMap[msgIndex], accumulatedAttrs...)
			if msgIndex > maxIndex {
				maxIndex = msgIndex
			}
			accumulatedAttrs = []types.Attribute{}
		}
		if len(authzAttrsMap) == 0 {
			// no authz_msg_index found, treat as special event and append it to all authz events
			// cannot use event directly since we need the filter for message.action
			specialEvents = append(specialEvents, types.StringEvent{
				Type:       event.Type,
				Attributes: accumulatedAttrs,
			})
		} else {
			for i, attrs := range authzAttrsMap {
				authzEvent := types.StringEvent{
					Type:       event.Type,
					Attributes: attrs,
				}
				authzEvents[i] = append(authzEvents[i], authzEvent)
			}
		}
	}
	eventsList := make(db.EventsList, maxIndex+1) // 0...maxIndex (inclusive) so +1
	for i := range eventsList {
		eventsList[i].Events = authzEvents[i]
		eventsList[i].Events = append(eventsList[i].Events, specialEvents...)
	}
	return eventsList, nil
}

func extractAuthzMessages(msg json.RawMessage) ([]json.RawMessage, error) {
	var msgExec struct {
		Msgs []json.RawMessage `json:"msgs"`
	}
	err := json.Unmarshal(msg, &msgExec)
	if err != nil {
		return nil, err
	}
	return msgExec.Msgs, nil
}

func EventContextFromAuthz(ctx db.EventContext, msgIndex int) (db.EventContext, error) {
	authzCtx := db.EventContext{
		Batch:       ctx.Batch,
		Timestamp:   ctx.Timestamp,
		TxHash:      ctx.TxHash,
		AuthzParent: &ctx,
	}
	var err error
	authzCtx.Messages, err = extractAuthzMessages(ctx.Messages[msgIndex])
	if err != nil {
		return db.EventContext{}, err
	}
	authzCtx.EventsList, err = extractAuthzEventsList(ctx.EventsList[msgIndex].Events)
	if err != nil {
		return db.EventContext{}, err
	}
	return authzCtx, nil
}

type Payload struct {
	db.EventContext
	MsgIndex int
}

func PayloadFromEventContext(ctx db.EventContext) *Payload {
	return &Payload{
		EventContext: ctx,
		MsgIndex:     -1,
	}
}

func (payload *Payload) Next() bool {
	payload.MsgIndex++
	// Not sure if len(payload.Messages) is always equal to len(payload.EventsList)
	// (e.g. in case of message execution failed?)
	// So check both to be safe
	return payload.MsgIndex < len(payload.Messages) && payload.MsgIndex < len(payload.EventsList)
}

func (payload *Payload) GetMessage() json.RawMessage {
	return payload.Messages[payload.MsgIndex]
}

func (payload *Payload) GetEvents() types.StringEvents {
	return payload.EventsList[payload.MsgIndex].Events
}

type EventProcessor func(payload *Payload, event *types.StringEvent) error

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

func (e *EventExtractor) runProcessors(payload *Payload, event *types.StringEvent, processors []EventProcessor) error {
	for _, processor := range processors {
		if err := processor(payload, event); err != nil {
			return err
		}
	}
	return nil
}

func (e *EventExtractor) extractTypeKeyValue(payload *Payload, event *types.StringEvent) error {
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

func (e *EventExtractor) extractTypeKey(payload *Payload, event *types.StringEvent) error {
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

func (e *EventExtractor) extractType(payload *Payload, event *types.StringEvent) error {
	err := e.runProcessors(payload, event, e.typeMap[event.Type])
	if err != nil {
		return err
	}
	return nil
}

func (e *EventExtractor) extractAll(payload *Payload) error {
	err := e.runProcessors(payload, nil, e.wildcards)
	if err != nil {
		return err
	}
	return nil
}

func (e *EventExtractor) Extract(ctx db.EventContext) error {
	payload := PayloadFromEventContext(ctx)
	for payload.Next() {
		events := payload.GetEvents()
		isAuthz := utils.GetEventsValue(events, "message", "action") == "/cosmos.authz.v1beta1.MsgExec"
		if isAuthz {
			authzCtx, err := EventContextFromAuthz(ctx, payload.MsgIndex)
			if err == nil {
				err = e.Extract(authzCtx)
				if err != nil {
					return err
				}
				continue
			}
			// TODO: ???
			logger.L.Warnw("failed to extract context from MsgExec and events", "ctx", ctx, "err", err)
		}
		for _, event := range events {
			type extractFuncType = func(*Payload, *types.StringEvent) error
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
	}
	return nil
}
