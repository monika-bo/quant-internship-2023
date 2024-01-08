package groups

type GroupRequest struct {
	Name     string   `json:"name"`
	Contacts []string `json:"contacts"`
}

func (g GroupRequest) ToModel() Group {
	return Group{

		Name:     g.Name,
		Contacts: g.Contacts,
	}
}
