package lang

import "github.com/code-game-project/cg-gen-events/cge"

type Generator interface {
	Generate(objects []cge.Object, gameName, dir string) error
}
