package lang

type Generator interface {
	Generate(gameName, dir string) error
}
