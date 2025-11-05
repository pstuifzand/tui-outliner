package template

import (
	"testing"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

func TestParseTypeSpec(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		spec    string
		wantErr bool
		kind    string
		values  []string
	}{
		{
			name: "enum type",
			key:  "status",
			spec: "enum|todo|in-progress|done",
			kind: "enum",
			values: []string{"todo", "in-progress", "done"},
		},
		{
			name: "number type",
			key:  "priority",
			spec: "number|1-5",
			kind: "number",
			values: []string{"1-5"},
		},
		{
			name: "date type",
			key:  "deadline",
			spec: "date",
			kind: "date",
			values: []string{},
		},
		{
			name: "string type",
			key:  "title",
			spec: "string",
			kind: "string",
		},
		{
			name:    "empty spec",
			spec:    "",
			wantErr: true,
		},
		{
			name:    "unknown kind",
			spec:    "unknown|value",
			wantErr: true,
		},
		{
			name:    "enum without values",
			spec:    "enum",
			wantErr: true,
		},
		{
			name:    "number with invalid range",
			spec:    "number|abc-def",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := ParseTypeSpec(tt.key, tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTypeSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if ts.Kind != tt.kind {
					t.Errorf("Kind = %v, want %v", ts.Kind, tt.kind)
				}
				if len(ts.Values) != len(tt.values) {
					t.Errorf("Values length = %v, want %v", len(ts.Values), len(tt.values))
				}
			}
		})
	}
}

func TestTypeSpecValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		value   string
		wantErr bool
	}{
		{
			name:    "enum valid",
			spec:    "enum|todo|done",
			value:   "todo",
			wantErr: false,
		},
		{
			name:    "enum invalid",
			spec:    "enum|todo|done",
			value:   "in-progress",
			wantErr: true,
		},
		{
			name:    "number valid",
			spec:    "number|1-5",
			value:   "3",
			wantErr: false,
		},
		{
			name:    "number too high",
			spec:    "number|1-5",
			value:   "10",
			wantErr: true,
		},
		{
			name:    "number not a number",
			spec:    "number|1-5",
			value:   "abc",
			wantErr: true,
		},
		{
			name:    "date valid",
			spec:    "date",
			value:   "2025-11-05",
			wantErr: false,
		},
		{
			name:    "date invalid format",
			spec:    "date",
			value:   "11/05/2025",
			wantErr: true,
		},
		{
			name:    "date invalid month",
			spec:    "date",
			value:   "2025-13-05",
			wantErr: true,
		},
		{
			name:    "date invalid day",
			spec:    "date",
			value:   "2025-02-30",
			wantErr: true,
		},
		{
			name:    "date valid leap year",
			spec:    "date",
			value:   "2024-02-29",
			wantErr: false,
		},
		{
			name:    "date invalid leap year",
			spec:    "date",
			value:   "2025-02-29",
			wantErr: true,
		},
		{
			name:    "string always valid",
			spec:    "string",
			value:   "anything goes",
			wantErr: false,
		},
		{
			name:    "list string valid",
			spec:    "list|string",
			value:   "apple, banana, orange",
			wantErr: false,
		},
		{
			name:    "list string empty item",
			spec:    "list|string",
			value:   "apple, , orange",
			wantErr: true,
		},
		{
			name:    "list enum valid",
			spec:    "list|enum|low|medium|high",
			value:   "low, high, medium",
			wantErr: false,
		},
		{
			name:    "list enum invalid item",
			spec:    "list|enum|low|medium|high",
			value:   "low, critical, medium",
			wantErr: true,
		},
		{
			name:    "list number valid",
			spec:    "list|number|1-5",
			value:   "1, 3, 5",
			wantErr: false,
		},
		{
			name:    "list number out of range",
			spec:    "list|number|1-5",
			value:   "1, 10, 3",
			wantErr: true,
		},
		{
			name:    "list number not a number",
			spec:    "list|number|1-5",
			value:   "1, abc, 3",
			wantErr: true,
		},
		{
			name:    "list date valid",
			spec:    "list|date",
			value:   "2025-01-01, 2024-12-31, 2025-02-14",
			wantErr: false,
		},
		{
			name:    "list date invalid",
			spec:    "list|date",
			value:   "2025-01-01, 2025-13-01, 2025-02-14",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, _ := ParseTypeSpec("test", tt.spec)
			err := ts.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTypeRegistry(t *testing.T) {
	tr := NewTypeRegistry()

	// Add types
	tr.AddType("status", "enum|todo|done")
	tr.AddType("priority", "number|1-5")

	// Validate
	if err := tr.Validate("status", "todo"); err != nil {
		t.Errorf("Valid status should not error: %v", err)
	}

	if err := tr.Validate("status", "invalid"); err == nil {
		t.Errorf("Invalid status should error")
	}

	// Get type
	ts := tr.GetType("status")
	if ts == nil {
		t.Errorf("GetType should return status type")
	}

	// Remove type
	tr.RemoveType("status")
	ts = tr.GetType("status")
	if ts != nil {
		t.Errorf("Removed type should return nil")
	}
}

func TestTypeRegistryOutlineIntegration(t *testing.T) {
	outline := model.NewOutline()
	tr := NewTypeRegistry()

	// Add types to registry
	tr.AddType("status", "enum|todo|done")
	tr.AddType("priority", "number|1-5")

	// Save to outline
	if err := tr.SaveToOutline(outline); err != nil {
		t.Fatalf("SaveToOutline failed: %v", err)
	}

	// Create new registry and load
	tr2 := NewTypeRegistry()
	if err := tr2.LoadFromOutline(outline); err != nil {
		t.Fatalf("LoadFromOutline failed: %v", err)
	}

	// Verify types were loaded
	if err := tr2.Validate("status", "todo"); err != nil {
		t.Errorf("Loaded type should validate: %v", err)
	}

	if err := tr2.Validate("priority", "3"); err != nil {
		t.Errorf("Loaded type should validate: %v", err)
	}
}

func TestValidateItem(t *testing.T) {
	tr := NewTypeRegistry()
	tr.AddType("status", "enum|todo|done")
	tr.AddType("priority", "number|1-5")

	item := model.NewItem("Test")
	item.Metadata.Attributes["status"] = "todo"
	item.Metadata.Attributes["priority"] = "3"

	// Valid item
	if err := tr.ValidateItem(item); err != nil {
		t.Errorf("Valid item should not error: %v", err)
	}

	// Invalid attribute
	item.Metadata.Attributes["status"] = "invalid"
	if err := tr.ValidateItem(item); err == nil {
		t.Errorf("Invalid item should error")
	}
}
