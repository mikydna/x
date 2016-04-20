package control

type Controlled interface {
	Start() error
	Stop() error
	Running() bool
}
