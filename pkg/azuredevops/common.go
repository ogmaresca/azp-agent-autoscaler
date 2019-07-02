package azuredevops

// Definition is the base type for Azure Devops responses
type Definition struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// User fields in the Azure Devops API
type User struct {
	DisplayName string `json:"displayName"`
	URL         string `json:"url"`
	ID          string `json:"id"`
	UniqueName  string `json:"uniqueName"`
	ImageURL    string `json:"imageUrl"`
	Descriptor  string `json:"descriptor"`
}

// Error is returned when an error occurs in the API, such as an invalid ID being used.
type Error struct {
	//ID             string `json:"$id"`
	//InnerException Error  `json:"innerException"`
	Message   string `json:"message"`
	TypeName  string `json:"typeName"`
	TypeKey   string `json:"typeKey"`
	ErrorCode int    `json:"errorCode"`
	EventID   int    `json:"eventId"`
}
