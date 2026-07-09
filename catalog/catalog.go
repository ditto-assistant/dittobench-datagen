// Package catalog defines the Ditto tool catalog presented to harnesses. Each
// entry has a name, a short description, and a JSON-schema parameter
// definition. This is the tool menu a harness sees on every RunRequest.
//
// Schemas mirror the production Ditto backend's tool surface (array-valued
// queries/urls/pairIds, goal-based dynamic workflows), so a harness graded
// here faces the same shapes it would face in the real product.
package catalog

import (
	"encoding/json"

	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// prop is one JSON-schema property: name, type, and description. Type
// "string[]" renders an array of strings; anything else is used verbatim
// ("string", "integer", "boolean").
type prop struct {
	name string
	typ  string
	desc string
}

// schema builds a JSON-schema object from ordered property definitions.
func schema(required []string, props ...prop) json.RawMessage {
	properties := map[string]any{}
	for _, p := range props {
		if p.typ == "string[]" {
			properties[p.name] = map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": p.desc,
			}
			continue
		}
		properties[p.name] = map[string]any{
			"type":        p.typ,
			"description": p.desc,
		}
	}
	s := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	b, _ := json.Marshal(s)
	return json.RawMessage(b)
}

// params keeps the string-only helper for simple tools.
func params(props map[string]string, required ...string) json.RawMessage {
	ps := make([]prop, 0, len(props))
	for name, desc := range props {
		ps = append(ps, prop{name: name, typ: "string", desc: desc})
	}
	return schema(required, ps...)
}

// Catalog returns the full Ditto tool catalog.
func Catalog() []protocol.ToolDefinition {
	return []protocol.ToolDefinition{
		{
			Name:        "create_image",
			Description: "Generate a new image from a text prompt.",
			Parameters:  params(map[string]string{"prompt": "what to draw"}, "prompt"),
		},
		{
			Name:        "edit_image",
			Description: "Edit an existing image given an instruction.",
			Parameters:  params(map[string]string{"image_url": "image to edit", "instruction": "how to change it"}, "image_url", "instruction"),
		},
		{
			Name:        "read_links",
			Description: "Read one or more URLs and return markdown text content.",
			Parameters: schema([]string{"urls"},
				prop{"urls", "string[]", "URLs to read"},
			),
		},
		{
			Name:        "search_web",
			Description: "Search live sources for one or more queries.",
			Parameters: schema([]string{"queries"},
				prop{"queries", "string[]", "Search queries"},
				prop{"num_results", "integer", "Number of results per query, default 10"},
				prop{"search_mode", "string", "Optional source filter, default all"},
			),
		},
		{
			Name:        "search_memories",
			Description: "Search past conversations and return compact memory summaries. Use fetch_memories for selected IDs that need full text. For named entities/topics, prefer search_subjects -> search_memories_in_subjects.",
			Parameters: schema([]string{"queries"},
				prop{"queries", "string[]", "Memory search queries"},
			),
		},
		{
			Name:        "search_subjects",
			Description: "Search the user's subject graph and return subject objects with id, name, description, and similarity.",
			Parameters: schema([]string{"queries"},
				prop{"queries", "string[]", "Subject search queries"},
			),
		},
		{
			Name:        "fetch_memories",
			Description: "Fetch full conversation text for selected memory pair IDs.",
			Parameters: schema([]string{"pairIds"},
				prop{"pairIds", "string[]", "Memory pair IDs to fetch"},
				prop{"stripImages", "boolean", "Exclude images, default true"},
			),
		},
		{
			Name:        "search_memories_in_subjects",
			Description: "Semantic memory search inside one subject. Use focused queries; fetch selected IDs for full text.",
			Parameters: schema([]string{"subject_id", "queries"},
				prop{"subject_id", "string", "Subject ID from search_subjects or the prompt"},
				prop{"queries", "string[]", "Focused queries within the subject"},
			),
		},
		{
			Name:        "artifacts",
			Description: "Create an interactive, previewable artifact (web app, doc, game).",
			Parameters:  params(map[string]string{"spec": "what to build"}, "spec"),
		},
		{
			Name:        "execute_agent_job",
			Description: "Dispatch a one-off background agent job. For a goal with clear independent parts, use execute_agent_workflow instead.",
			Parameters:  params(map[string]string{"task": "the task to run"}, "task"),
		},
		{
			Name:        "execute_agent_workflow",
			Description: "Plan a complex task as multiple parallel sub-agents. Use only when the task has clear independent parts (multi-angle research, audits across dimensions); for single one-shot tasks use execute_agent_job instead.",
			Parameters: schema([]string{"goal"},
				prop{"goal", "string", "High-level intent; the planner decomposes this into 2-6 parallel sub-tasks."},
				prop{"max_parallel", "integer", "Cap on concurrent workers (default 4, max 8)."},
			),
		},
		{
			Name:        "get_agent_job_status",
			Description: "Check the status of a running or finished agent job.",
			Parameters:  params(map[string]string{"job_id": "the job ID"}, "job_id"),
		},
		{
			Name:        "list_agent_jobs",
			Description: "List the user's recent agent jobs.",
			Parameters: schema(nil,
				prop{"status", "string", "Filter by status: pending, running, completed, failed, cancelled (optional)"},
				prop{"limit", "integer", "Max results to return (default 20, max 50)"},
			),
		},
		{
			Name:        "file_feedback_for_team",
			Description: "Send feedback or a bug report to the Ditto team.",
			Parameters:  params(map[string]string{"message": "the feedback message"}, "message"),
		},
		{
			Name:        "set_theme",
			Description: "Change the app's color theme.",
			Parameters:  params(map[string]string{"theme": "theme name, e.g. dark or light"}, "theme"),
		},
		{
			Name:        "set_main_model",
			Description: "Change the primary chat model.",
			Parameters:  params(map[string]string{"model": "model identifier"}, "model"),
		},
		{
			Name:        "set_reasoning_effort",
			Description: "Set the reasoning effort level for responses.",
			Parameters:  params(map[string]string{"effort": "low, medium, or high"}, "effort"),
		},
		{
			Name:        "set_chat_tool_preferences",
			Description: "Enable or disable specific tools for chat.",
			Parameters:  params(map[string]string{"preferences": "tool preference settings"}, "preferences"),
		},
	}
}
