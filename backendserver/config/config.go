package bconfig

//Configuration struct holds information of the configured backendservers
type BackendConfiguration struct {
	Servers []struct {
		Host   string `json:"Host"`
		Port   string `json:"Port"`
		Weight int    `json:"Weight"`
	} `json:"Servers"`
}
