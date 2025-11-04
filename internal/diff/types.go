package diff

// ItemData represents parsed diff format data for a single item
type ItemData struct {
	ID         string
	Text       string
	ParentID   string
	Position   int
	Tags       []string
	Attributes map[string]string
	Created    string
	Modified   string
}

// DiffResult contains the analysis of changes between two outlines
type DiffResult struct {
	NewItems      map[string]*ItemData
	DeletedItems  map[string]*ItemData
	ModifiedItems map[string]*ItemChange
}

// ItemChange describes what changed for an item
type ItemChange struct {
	Item             *ItemData
	OldItem          *ItemData
	TextChanged      bool
	OldText          string
	StructureChanged bool
	OldParentID      string
	OldPosition      int
	TagsAdded        []string
	TagsRemoved      []string
	AttrsAdded       map[string]string
	AttrsRemoved     map[string]string
	AttrsChanged     map[string][2]string // attrName -> [oldValue, newValue]
	ModifiedChanged  bool
	OldModified      string
}

// DiffLineType indicates the type of diff line for rendering
type DiffLineType int

const (
	DiffTypeHeader DiffLineType = iota
	DiffTypeNewSection
	DiffTypeDeletedSection
	DiffTypeModifiedSection
	DiffTypeNewItem
	DiffTypeDeletedItem
	DiffTypeModifiedItem
	DiffTypeItemDetail
	DiffTypeSummary
	DiffTypeBlank
)

// DiffLine represents a rendered line in diff output
type DiffLine struct {
	Type       DiffLineType
	Content    string
	Indent     int  // Indentation level
	IsCollapse bool // Whether this starts a collapsible section
}
