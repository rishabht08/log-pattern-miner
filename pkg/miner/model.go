package miner

type Config struct {
	RegexMap map[string]string `json:"regex_map"`
}

type MinedTemplate struct {
	OriginalLog string   `json:"original_log"`
	Template    string   `json:"template"`
	TemplateID  string   `json:"template_id"`
	Parameters  []string `json:"parameters"`
	ParamID     string   `json:"param_id"`
	Tokens      []string `json:"tokens"`
}

type SerializableNode struct {
	Children map[string]*SerializableNode `msgpack:"children"`
}

type SerializablePatternTree struct {
	Root *SerializableNode `msgpack:"root"`
}
