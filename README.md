# beecs_ecotox


Work-in-progress re-implementation of the [BEEHAVE_ecotox](https://academic.oup.com/etc/article/41/11/2870/7730717) model
in [Go](https://go.dev) using the [Arche](https://github.com/mlange-42/arche) Entity Component System (ECS).

All the hard work to develop, parameterize and validate the original BEEHAVE model was done by Dr. Matthias Becher and co-workers.
The hard work of developing the _ecotox-addition was done by Thomas G. Preuss, Benoit Goussen and their coworkers.
Martin Lange at the UFZ Leipzig re-implemented the original BEEHAVE as a GO-ECS-model and put in the creativity and work necessary to create a functioning model.
I was not involved in that development in any way. I am just re-implementing the _ecotox-Version following its ODD Protocol and the NetLogo code on top of the already re-implemented model of Martin Lange.

Beecs_ecotox is currently still in developement and needs to be verified. It is still being debugged and tested as of now and cannot recreate the exact results as the original model in all scenarios yet. Tendencies already show very similar qualitative behaviour, though. The water-foraging module is not implemented as of now and might never be, as there appears to be a lack of relevance in doing so.

ToDo´s for now:
- further testing to find out why there are deviances in foraging behavious that lead to:
    - different amounts of contact exposition of foragers
    - different amounts of PPP being taken into hive, especially pollen 
- introduce the option to sim multiple pesticides as in the BEEHAVE_ecotox-update 2023



## Usage

### Command line app

no current implementation here

### Graphical user interface

no current implementation here

### Library

To add beecs as a dependency to an existing Go project, run this in the project's root folder:

```
go get github.com/fzeitner/beecs_ecotox_test

```

