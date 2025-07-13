// Re-implementation of the [BEEHAVE_ecotox] model
// in [Go] using the [Arche] Entity Component System (ECS).
//
// Most important packages:
//
//   - [github.com/fzeitner/beecs_ecotox/comp] -- Components
//   - [github.com/fzeitner/beecs_ecotox/sys] -- Systems / sub-models
//   - [github.com/fzeitner/beecs_ecotox/params] -- Model parameters
//   - [github.com/fzeitner/beecs_ecotox/globals] -- Global state variables
//
// [Go]: https://go.dev
// [Arche]: https://github.com/mlange-42/arche
//
// [BEEHAVE]: https://beehave-model.net
package beecs
