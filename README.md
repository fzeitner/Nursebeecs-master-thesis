# Nursebeecs-master-thesis


This is the prototype of the nursebeecs model that was created within the scope of my master thesis. It is the product of a large scale literature research to explicitly implement nurse bees and the nursing task into the in-hive colony structure of the beecs/BEEHAVE model. A few of the changes are rather experimental, and it is possible that an adjusted version of this model will be published scientifically in the future. For now, this project serves as documentation of my master thesis and can be used to explore the nurse bee related consumption dynamics and its effects on simulated pesticide exposure.

Nursebeecs is based on the [beecs](https://github.com/mlange-42/beecs) model and incorporates the mechanistic effect modules of the [BEEHAVE_ecotox](https://academic.oup.com/etc/article/41/11/2870/7730717) model (which created beecs_ecotox, which will be linked here once it is cleanly implemented as a beecs fork) in [Go](https://go.dev) using the [Ark](https://github.com/mlange-42/ark) Entity Component System (ECS).

All the hard work to develop, parameterize, and validate the original BEEHAVE model was done by Dr. Matthias Becher and co-workers.
The development of the _ecotox-addition was done by Thomas G. Preuss, Benoit Goussen, Matthias Becher and their co-workers.
Martin Lange at the UFZ Leipzig re-implemented the original BEEHAVE as a ECS-based model in the Go programming language called [beecs](https://github.com/mlange-42/beecs) and put in the creativity and work necessary to create a well-functioning model.
I was not involved in any of these development processes, but I re-implemented the _ecotox-changes following the ODD Protocol and the NetLogo code on top of the already re-implemented model of M. Lange and used this as a baseline to further develope a model for my master thesis.

Nursebeecs is an exploratory, nurse bee centered addition to the beecs/BEEHAVE model family with the goal of modeling in-hive population dynamics more accurately and estimating the exposure of nurse bees to pesticides. This enables a more realistic estimation of lethal effects and the exploration of sublethal effects on brood care. The primary use of this repository is to explore the model adaptations and create results for my master thesis at the Osnabrück University. 


## open ToDo´s still coming:
- re-implement this cleanly and in a way that ensures compatibility to M. Lange's [beecs-cli](https://github.com/mlange-42/beecs-cli) and [beecs-ui](https://github.com/mlange-42/beecs-ui), which allow much more intuitive controls of model simulations
- reintroduce tests
- fix the floating point error that (extremely rarely, but annoyingly) occurs during comparison of the honey stores to the etox_honey stores (introduced by BEEHAVE_ecotox) and triggers a panic
- (optionally) increase model performance for better runtimes


## Usage

### Clone this repository and use of main.go-files

As of now, the only safe and tested way to run simulations and create results is via creation of a dedicated "main.go" file: 
- choose a model version ("model.Default", "model.DefaultEtox", "model.DefaultNbeecs", "model.DefaultNbeecsEtox")
- create the Default parameters and (optionally) change these
- add observers with the metrics you want to observe (observers will write CSV files with the metrics that can be aggregated and visualized)
- run the model
- (aggregate and) visualize the data

The easiest way to understand this process is to check the [examples](https://github.com/fzeitner/Nursebeecs-master-thesis/tree/main/_examples), which include the original examples of M. Lange for basic beecs functions and two additional folders with examples that illustrate the controls for beecs_ecotox and nursebeecs. Aside from this, you can also check some of the existing main.go files in the [etox_validation_testing](https://github.com/fzeitner/Nursebeecs-master-thesis/tree/main/etox_validation_testing) or [nursebeecs_testing](https://github.com/fzeitner/Nursebeecs-master-thesis/tree/main/nursebeecs_testing) folders to see my application of these models to create results.

This project will not change in content anymore, but will be cleaned up tremendously to enable more intuitive use for people, who were not originally involved in my thesis or this project. It might be re-implemented more cleanly as a fork of M. Lange's [beecs](https://github.com/mlange-42/beecs) in the future and it might still receive support/compatibility with [beecs-cli](https://github.com/mlange-42/beecs-cli) and [beecs-ui](https://github.com/mlange-42/beecs-ui) during this process or before.

### Command line app

no current implementation here

### Graphical user interface

no current implementation here



### Installation/Cloning

To use this repository for simulations, it is currently advised to first create a local copy via

```
git clone https://github.com/fzeitner/Nursebeecs-master-thesis

```
