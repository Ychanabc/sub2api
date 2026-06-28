package service

import (
	"encoding/json"
	"strings"

	"github.com/tidwall/gjson"
)

const conversationAuditTextLimit = 12000

func TruncateConversationAuditText(text string) string {
	text = strings.TrimSpace(text)
	if len(text) <= conversationAuditTextLimit {
		return text
	}
	return text[:conversationAuditTextLimit]
}

// extractAnthropicResponseText 从非流式 Anthropic 响应体中提取
// content[].text 文本，用于对话审计。解析失败时返回空串。
func extractAnthropicResponseText(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var builder strings.Builder
	gjson.GetBytes(body, "content").ForEach(func(_, item gjson.Result) bool {
		if item.Get("type").String() == "text" {
			if txt := item.Get("text").String(); txt != "" {
				if builder.Len() > 0 {
					builder.WriteString("\n")
				}
				builder.WriteString(txt)
			}
		}
		return builder.Len() < conversationAuditTextLimit
	})
	if builder.Len() > 0 {
		return builder.String()
	}
	// 回退：OpenAI 兼容格式 choices[].message.content
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(body, &parsed); err == nil {
		if c := gjson.GetBytes(body, "choices.0.message.content"); c.Exists() {
			return c.String()
		}
	}
	return ""
}

// extractOpenAINonStreamResponseText 从 OpenAI 响应体（Responses / Chat Completions）
// 中提取助手输出文本，用于对话审计。解析失败返回空串。
func extractOpenAINonStreamResponseText(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var builder strings.Builder
	appendText := func(txt string) bool {
		if txt == "" {
			return true
		}
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(txt)
		return builder.Len() < conversationAuditTextLimit
	}

	// Responses API: output[].content[].text (type == output_text)
	gjson.GetBytes(body, "output").ForEach(func(_, item gjson.Result) bool {
		if item.Get("type").String() != "message" {
			return true
		}
		cont := true
		item.Get("content").ForEach(func(_, part gjson.Result) bool {
			if part.Get("type").String() == "output_text" {
				cont = appendText(part.Get("text").String())
			}
			return cont
		})
		return cont
	})
	if builder.Len() > 0 {
		return builder.String()
	}

	// Chat Completions: choices[].message.content
	if c := gjson.GetBytes(body, "choices.0.message.content"); c.Exists() {
		if c.Type == gjson.String {
			return c.String()
		}
		// content 数组形式
		c.ForEach(func(_, part gjson.Result) bool {
			return appendText(part.Get("text").String())
		})
		if builder.Len() > 0 {
			return builder.String()
		}
	}

	// Responses API top-level output_text 便捷字段
	if ot := gjson.GetBytes(body, "output_text"); ot.Exists() && ot.Type == gjson.String {
		return ot.String()
	}
	return ""
}
