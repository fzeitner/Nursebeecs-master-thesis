# Validation

Validation of the beecs implementation against the original BEEHAVE_ecotox model.

The plots below show simulation results for one year. Lines are medians of 100 simulations, areas are 5% and 95% quantiles.

This test uses the wheather from Rothamsted (2009) to explore the effects of fenoxycarb. The weather setting can be adequately simulated as is evident in this test "Rothamsted2009_noPPP". The effects of fenoxycarb in this example cannot be reproduces adequately though. The change of the application day to a few days later could also not change the effect on the colony qualitatively. The differences appear to take root in the different concentrations of PPP in the pollen stores inhive, which I assume is caused by differencs in foraging behaviour on the day of application again.
Weirdly the Larvae do react exactly the same way in both model versions though, the differences only stem from IHbee/forager dynamcs. Concretely, more IHbees seem to convert to foragers earlier in the beecs version than in the netlogo version of the model.

Still have to debug exactly where the difference in all behaviours are caused and do some more testing.

## Colony structure

- gonna find out how to put plots here to show similarities once I actually make this repo public in one way or another


## Honey and pollen stores

- gonna find out how to put plots here to show similarities once I actually make this repo public in one way or another

