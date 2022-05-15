package lang

import "github.com/code-game-project/cg-gen-events/cge"

type Generator interface {
	Generate(server bool, metadata cge.Metadata, objects []cge.Object, dir string) error
}
