# beecs_masterthesis


Work-in-progress on top of the re-implementation of the [BEEHAVE_ecotox](https://academic.oup.com/etc/article/41/11/2870/7730717) model
in [Go](https://go.dev) using the [Arche](https://github.com/mlange-42/arche) Entity Component System (ECS).

All the hard work to develop, parameterize and validate the original BEEHAVE model was done by Dr. Matthias Becher and co-workers.
The hard work of developing the _ecotox-addition was done by Thomas G. Preuss, Benoit Goussen and their coworkers.
Martin Lange at the UFZ Leipzig re-implemented the original BEEHAVE as a GO-ECS-model and put in the creativity and work necessary to create a functioning model.
I was not involved in that development in any way. I just re-implemented the _ecotox-Version following its ODD Protocol and the NetLogo code on top of the already re-implemented model of Martin Lange and now use this as a baseline to further develope a model for my master thesis.

Beecs_masterthesis is currently still in developement with the current goal of modelling nurse bee exposure and estimating lethal and sublethal effects. My primary reason for creating this is to use this for my master thesis at the University Osnabrueck. 
This model is still being debugged and tested as of now.

ToDo´s for now:
- implement beeGUTS
- implement nurse bee cohort
- lots and LOTS of testing and validating



## Usage

### Command line app

no current implementation here

### Graphical user interface

no current implementation here

### Library

To add this as a dependency to an existing Go project, run this in the project's root folder:

```
go get github.com/fzeitner/beecs_masterthesis

```

