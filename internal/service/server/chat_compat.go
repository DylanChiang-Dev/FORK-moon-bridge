package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"moonbridge/internal/protocol/chat"
	"moonbridge/internal/protocol/openai"
)

// handleChatCompletions accepts OpenAI Chat Completions API requests and
// converts them into the Responses API format before delegating to the
// existing handleResponses pipeline.
//
// Conversion overview:
//   ChatRequest{model,messages,tools,stream,...} → ResponsesRequest{model,input,tools,stream,...}
// Messages are packed into the Responses "input" array as user/assistant/system items.
func (server *Server) handleChatCompletions(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeOpenAIError(writer, http.StatusMethodNotAllowed, openai.ErrorResponse{Error: openai.ErrorObject{
			Message: "method not allowed",
			Type:    "invalid_request_error",
			Code:    "method_not_allowed",
		}})
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, openai.ErrorResponse{Error: openai.ErrorObject{
			Message: "failed to read request body",
			Type:    "invalid_request_error",
			Code:    "invalid_request_body",
		}})
		return
	}

	var chatReq chat.ChatRequest
	if err := json.Unmarshal(body, &chatReq); err != nil {
		writeOpenAIError(writer, http.StatusBadRequest, openai.ErrorResponse{Error: openai.ErrorObject{
			Message: "invalid JSON request body",
			Type:    "invalid_request_error",
			Code:    "invalid_json",
		}})
		return
	}

	// Convert Chat Completions request to Responses API request.
	respReq := chatToResponsesRequest(chatReq)

	respBody, err := json.Marshal(respReq)
	if err != nil {
		writeOpenAIError(writer, http.StatusInternalServerError, openai.ErrorResponse{Error: openai.ErrorObject{
			Message: "failed to marshal converted request",
			Type:    "server_error",
			Code:    "internal_error",
		}})
		return
	}

	// Replace the request body with the converted Responses JSON and delegate.
	request.Body = io.NopCloser(bytes.NewReader(respBody))
	server.handleResponses(writer, request)
}

// chatToResponsesRequest converts an OpenAI Chat Completions request into
// a Moon Bridge Responses API request.
func chatToResponsesRequest(chatReq chat.ChatRequest) openai.ResponsesRequest {
	rr := openai.ResponsesRequest{
		Model:  chatReq.Model,
		Stream: chatReq.Stream,
	}

	if chatReq.MaxTokens > 0 {
		rr.MaxOutputTokens = chatReq.MaxTokens
	}
	if chatReq.Temperature != nil {
		rr.Temperature = chatReq.Temperature
	}
	if chatReq.TopP != nil {
		rr.TopP = chatReq.TopP
	}

	// Convert messages → Responses input array.
	inputItems := make([]json.RawMessage, 0, len(chatReq.Messages))
	for _, msg := range chatReq.Messages {
		item := chatMessageToInputItem(msg)
		if item != nil {
			inputItems = append(inputItems, *item)
		}
	}
	if len(inputItems) > 0 {
		inputJSON, _ := json.Marshal(inputItems)
		rr.Input = inputJSON
	}

	// Convert tools.
	if len(chatReq.Tools) > 0 {
		rr.Tools = make([]openai.Tool, 0, len(chatReq.Tools))
		for _, t := range chatReq.Tools {
			rr.Tools = append(rr.Tools, openai.Tool{
				Type:        t.Type,
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			})
		}
	}

	// Tool choice.
	if len(chatReq.ToolChoice) > 0 {
		rr.ToolChoice = chatReq.ToolChoice
	}

	return rr
}

// chatMessageToInputItem converts a single Chat Completions message into a
// Responses API input item (raw JSON). Returns nil for unsupported roles.
func chatMessageToInputItem(msg chat.ChatMessage) *json.RawMessage {
	var result json.RawMessage

	switch msg.Role {
	case "system":
		content := chatContentToString(msg.Content)
		result, _ = json.Marshal(map[string]string{
			"role":    "system",
			"content": content,
		})
	case "user":
		result = chatUserMessageToInput(msg)
	case "assistant":
		result = chatAssistantMessageToInput(msg)
	case "tool":
		result = chatToolMessageToInput(msg)
	default:
		return nil
	}

	if len(result) == 0 {
		return nil
	}
	return &result
}

func chatContentToString(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, part := range v {
			if m, ok := part.(map[string]any); ok {
				if t, ok := m["text"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		data, _ := json.Marshal(content)
		return string(data)
	}
}

func chatUserMessageToInput(msg chat.ChatMessage) json.RawMessage {
	switch v := msg.Content.(type) {
	case string:
		result, _ := json.Marshal(map[string]any{
			"role":    "user",
			"content": v,
		})
		return result
	case []any:
		parts := make([]map[string]any, 0, len(v))
		for _, part := range v {
			if m, ok := part.(map[string]any); ok {
				partType, _ := m["type"].(string)
				switch partType {
				case "text":
					if text, ok := m["text"].(string); ok {
						parts = append(parts, map[string]any{
							"type": "input_text",
							"text": text,
						})
					}
				case "image_url":
					if img, ok := m["image_url"].(map[string]any); ok {
						if url, ok := img["url"].(string); ok {
							parts = append(parts, map[string]any{
								"type":      "input_image",
								"image_url": url,
							})
						}
					}
				}
			}
		}
		result, _ := json.Marshal(map[string]any{
			"role":    "user",
			"content": parts,
		})
		return result
	default:
		data, _ := json.Marshal(msg.Content)
		result, _ := json.Marshal(map[string]any{
			"role":    "user",
			"content": string(data),
		})
		return result
	}
}

func chatAssistantMessageToInput(msg chat.ChatMessage) json.RawMessage {
	item := map[string]any{
		"role": "assistant",
	}
	if text := chatContentToString(msg.Content); text != "" {
		item["content"] = text
	}
	if len(msg.ToolCalls) > 0 {
		fcItems := make([]map[string]any, 0, len(msg.ToolCalls))
		for _, tc := range msg.ToolCalls {
			if tc.Type != "function" {
				continue
			}
			args := tc.Function.Arguments
			if len(args) == 0 {
				args = json.RawMessage("{}")
			}
			fcItems = append(fcItems, map[string]any{
				"type":      "function_call",
				"id":        tc.ID,
				"call_id":   tc.ID,
				"name":      tc.Function.Name,
				"arguments": args,
			})
		}
		if len(fcItems) > 0 {
			data, _ := json.Marshal(fcItems)
			item["content"] = json.RawMessage(data)
		}
	}
	result, _ := json.Marshal(item)
	return result
}

func chatToolMessageToInput(msg chat.ChatMessage) json.RawMessage {
	result, _ := json.Marshal(map[string]any{
		"type":    "function_call_output",
		"call_id": msg.ToolCallID,
		"output":  chatContentToString(msg.Content),
	})
	return result
}

