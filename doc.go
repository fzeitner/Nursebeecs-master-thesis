// Re-implementation of the [BEEHAVE_ecotox] model - with additions of exploratory nurse bee dynamics -
// in [Go] using the [Ark] Entity Component System (ECS). The basis for this is the [beecs] model of M. Lange
//
// Most important packages:
//
//   - [github.com/fzeitner/Nursebeecs-master-thesis/tree/main/sys] -- Systems / submodels of original beecs
//   - [github.com/fzeitner/Nursebeecs-master-thesis/tree/main/sys_etox] -- Systems / submodels of the ecotox and nurse bee centered additions
//   - [github.com/fzeitner/Nursebeecs-master-thesis/tree/main/params] -- Model parameters of original beecs
//   - [github.com/fzeitner/Nursebeecs-master-thesis/tree/main/params_etox] -- Model parameters of the ecotox and nurse bee centered additions
//   - [github.com/fzeitner/Nursebeecs-master-thesis/tree/main/globals] -- Global state variables of original beecs
//   - [github.com/fzeitner/Nursebeecs-master-thesis/tree/main/globals_etox] -- Global state variables of the ecotox and nurse bee centered additions
//
// [Go]: https://go.dev
// [Ark]: https://github.com/mlange-42/ark
// [BEEHAVE_ecotox]: https://github.com/ibacon-GmbH-Modelling/BEEHAVEecotox
// [beecs]: https://github.com/mlange-42/beecs
//
// [BEEHAVE]: https://beehave-model.net
package Nursebeecs
