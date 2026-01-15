package backend

type ServerPool struct {
	Backends []*Backend `json:"backends"`
	Current uint64 `json:"current"` // Used for Round-Robin
}

func (sp *ServerPool) do() {
	
}