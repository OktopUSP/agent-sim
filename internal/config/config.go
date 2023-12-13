package config

type Config struct {
	SimNumber    int
	NumToStartId int
	Prefix       string
	Address      string
	Port         string
	Mtp          string
	Path         string
}

func NewConfig(
	simNumber int,
	numToStartId int,
	prefix string,
	addr string,
	port string,
	mtp string,
	path string,
) Config {
	return Config{
		SimNumber:    simNumber,
		NumToStartId: numToStartId,
		Prefix:       prefix,
		Address:      addr,
		Port:         port,
		Mtp:          mtp,
		Path:         path,
	}
}
