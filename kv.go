package goblet

type KvDriver interface {
	Get(name string)
}
