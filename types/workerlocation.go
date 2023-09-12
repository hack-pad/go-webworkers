package types

type WorkerLocation struct {
	Hash     string
	Host     string
	HostName string
	Href     string
	Origin   string
	PathName string
	Port     string
	Protocol string
	Search   string
}

func (loc WorkerLocation) String() string {
	return loc.Href
}
