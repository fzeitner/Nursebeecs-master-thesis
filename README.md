# Nursebeecs-master-thesis


Work-in-progress on top of the re-implementation of the [BEEHAVE_ecotox](https://academic.oup.com/etc/article/41/11/2870/7730717) model
in [Go](https://go.dev) using the [Ark](https://github.com/mlange-42/ark) Entity Component System (ECS).

All the hard work to develop, parameterize and validate the original BEEHAVE model was done by Dr. Matthias Becher and co-workers.
The work of developing the _ecotox-addition was done by Thomas G. Preuss, Benoit Goussen and their coworkers.
Martin Lange at the UFZ Leipzig re-implemented the original BEEHAVE as a GO-ECS-model called [beecs](https://github.com/mlange-42/beecs) and put in the creativity and work necessary to create a functioning model.
I was not involved in that original development, but I re-implemented the _ecotox-Version following its ODD Protocol and the NetLogo code on top of the already re-implemented model of Martin Lange and now use this as a baseline to further develope a model for my master thesis.

Nursebeecs is an exploratory, nurse-bee-centered addition to the beecs/BEEHAVE model family with goal of modeling the population dynamics more accurately and estimating the exposure of nurse bees to pesticides. Also, this enables the estimation of lethal and brood-care-focused sublethal effects. The primary reason use of this repository is to explore the model adaptations and create results for my master thesis at the University Osnabrueck. 


## open ToDo´s still coming:
- clean up the code: add some comments, delete old code fragments, rename some variables (to ensure adherence to naming conventions), improve runtime performance with more efficient code, ...
- decouple nursing-related parameters from etox-related parameters for clarity 
- re-implement this cleanly and in a way that ensures compatibility to M. Lange's [beecs-cli](https://github.com/mlange-42/beecs-cli) and [beecs-ui](https://github.com/mlange-42/beecs-ui), which allow much more intuitive controls of model simulations
- reintroduce tests
- fix the floating point error that (very rarely) occurs during comparison of the honey stores to the etox_honey stores (introduced by BEEHAVEecotox) and triggers a panic


## Usage

### Clone this repository and use of main.go-files

As of now, the only safe and tested way to run simulations and create results is to create a main.go file, choose a model version ("model.Default", "model.Default_nbeecs", "model_etox.Default", "model_etox.Default_nbeecs"), create the Default parameters and (optionally) change these, add observers with the metrics you want to observe and run the model. The observers will write CSV files with the metrics that can be aggregated and visualized. The easiest way to understand this process is to check some of the existing main.go files in the [etox_validation_testing](https://github.com/fzeitner/Nursebeecs-master-thesis/tree/main/etox_validation_testing) or [nursebeecs_testing](https://github.com/fzeitner/Nursebeecs-master-thesis/tree/main/nursebeecs_testing) folders. Note that - as of now - running the "Default_nbeecs" model versions requires to activate the boolean parameter "NewConsumption" of the Nursing-parameter ("params_etox.Nursing"), because in the Nursbeecs model the honey and pollen consumption subsystems have been replaced by one NurseConsumption subsystem, which regulates both. This will still be patched to be much more user friendly and clear. This complete project will not change its content anymore, but will be cleaned up tremendously to enable much more intuitive use for everyone not originally involved in the thesis.

### Command line app

no current implementation here

### Graphical user interface

no current implementation here



### Installation

To use this repository for simulations, it is currently necessary first create a local copy via

```
git clone https://github.com/fzeitner/Nursebeecs-master-thesis

```
