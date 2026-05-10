package editRecord

type editPayload struct {
	Name      string   `json:"name"`
	Sensitive bool     `json:"sensitive"`
	Types     []string `json:"types"`
	DateOfDoc string   `json:"dateOfDoc"`
}
