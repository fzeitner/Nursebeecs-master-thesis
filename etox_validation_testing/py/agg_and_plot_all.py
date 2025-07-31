from os import path

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

def agg_beecs(file_pattern, out_file):
    data = None

    idx = 0
    while True:
        file = file_pattern % (idx,)
        if not path.exists(file):
            break

        run = pd.read_csv(file, delimiter=";")
        run = run.rename(columns={"t": "ticks"})
        run.insert(1, "run", idx)
        if data is None:
            data = run
        else:
            data = pd.concat([data, run])

        idx += 1

    runs = pd.unique(data.run)
    runs.sort()
    ticks = pd.unique(data.ticks)
    ticks.sort()

    columns = list(data.columns)[2:]

    out = pd.DataFrame(data={"ticks": ticks}, index=ticks)

    for column in columns:
        cols = [
            column + "_Q05",
            column + "_Q10",
            column + "_Q25",
            column + "_Q50",
            column + "_Q75",
            column + "_Q90",
            column + "_Q95",
        ]
        for col in cols:
            out[col] = 0.
        for tick in ticks:
            values = data[column][data.ticks == tick]

            inf = np.isinf(values)
            if sum(inf) > 0:
                _ = 1

            q = np.quantile(values, [0.05, 0.1, 0.25, 0.5, 0.75, 0.9, 0.95])
            out.loc[tick, cols] = q

    out.to_csv(out_file, sep=";", index=False)


def agg_netlogo(file, out_file):
    data = pd.read_csv(file, delimiter=",", header=6)

    runs = pd.unique(data.RAND_SEED)
    runs.sort()
    ticks = pd.unique(data.ticks)
    ticks.sort()

    columns = list(data.columns)[-19:]

    out = pd.DataFrame(data={"ticks": ticks}, index=ticks)

    for column in columns:
        cols = [
            column + "_Q05",
            column + "_Q10",
            column + "_Q25",
            column + "_Q50",
            column + "_Q75",
            column + "_Q90",
            column + "_Q95",
        ]
        for col in cols:
            out[col] = 0.
        for tick in ticks:
            values = data[column][data.ticks == tick]
            q = np.quantile(values, [0.05, 0.1, 0.25, 0.5, 0.75, 0.9, 0.95])
            out.loc[tick, cols] = q

    out.to_csv(out_file, sep=";", index=False)


def plot_quantiles(netlogo_file, beecs_file, out_dir, format, appday):
    data_beehave = pd.read_csv(netlogo_file, sep=r'\s*;\s*', engine='python')
    data_beecs = pd.read_csv(beecs_file, sep=r'\s*;\s*', engine='python')

    columns = list(data_beehave.columns)[1:]
    columns = pd.unique(pd.Series(c[:-4] for c in columns))
    quantiles = [
        ("Q05", 5),
        ("Q10", 10),
        ("Q25", 25),
        ("Q50", 50),
        ("Q75", 75),
        ("Q90", 90),
        ("Q95", 95),
    ]

    for col in columns:
        plot_column(
            data_beehave,
            data_beecs,
            col,
            quantiles,
            path.join(out_dir, col + "." + format),
            appday
        )


def plot_column(data_beehave, data_beecs, column, quantiles, image_file, appday):
    median_col = quantiles[len(quantiles) // 2][0]

    fig, ax = plt.subplots(figsize=(10, 4))
    for data, col, model in [
        (data_beecs, "blue", "beecs"),
        (data_beehave, "red", "BEEHAVE"),
    ]:      
        q10 = data[column + "_Q05"]
        q90 = data[column + "_Q95"]
        q50 = data[column + "_Q50"]

        ax.plot(data.ticks, q50, c=col, label=model)
        ax.fill_between(data.ticks, q10, q90, color=col, alpha=0.1)

    ax.set_title(column)
    ax.set_xlabel("time [d]", fontsize="12")
    if appday > 0:
        ax.vlines(appday, 0, max(q90), linestyle = "--", color = "gray", label = "application day")   # have to change the appday manually in func
    ax.legend()
    fig.tight_layout()

    plt.savefig(image_file)
    plt.close()



if __name__ == "__main__":
    ### change test folder and day of application manually here, applicationday is only relevant for 
    ### adding a visual indicator in plots, does not change anything regarding the results
    appdays = {"default_etox" : 0,                     # appday = 0 for no application
              "default_dimethoate": 189, 
              "default_dimethoate_GUTS": 189, 
              "etox_tunnel_control": 0,
              "etox_tunnel_dimethoate": 217, 
              "Rothamsted2009_fenoxycarb": 189, 
              "Rothamsted2009_noPPP": 0,
    }
    testfolders = ["default_etox", "default_dimethoate", "default_dimethoate_GUTS", "etox_tunnel_control",
                   "etox_tunnel_dimethoate", "Rothamsted2009_fenoxycarb", "Rothamsted2009_noPPP", ]
    folder = testfolders[2]



    run_all = False                   # True if you want to create all plots at once, just make sure to have run the sims beforehand. netlogo.csv's are provided
    agg_all = False
    agg_net = False
    agg_bee = True

    if run_all:
        for folder in testfolders:
            if agg_all:
                agg_beecs("etox_validation_testing/" + folder + "/out/beecs-%04d.csv", "etox_validation_testing/"+ folder +"/beecs.csv")
                agg_netlogo("etox_validation_testing/" + folder + "/out/netlogo.csv", "etox_validation_testing/" + folder + "/netlogo.csv")
            elif agg_net:
                agg_netlogo("etox_validation_testing/" + folder + "/out/netlogo.csv", "etox_validation_testing/" + folder + "/netlogo.csv")
            elif agg_bee:
                agg_beecs("etox_validation_testing/" + folder + "/out/beecs-%04d.csv", "etox_validation_testing/"+ folder +"/beecs.csv")


            plot_quantiles(
                "etox_validation_testing/" + folder + "/netlogo.csv",
                "etox_validation_testing/" + folder + "/beecs.csv",
                "etox_validation_testing/" + folder ,
                #"png",
                "svg",
                appdays[folder],
            )
    else:
        if agg_all:
                agg_beecs("etox_validation_testing/" + folder + "/out/beecs-%04d.csv", "etox_validation_testing/"+ folder +"/beecs.csv")
                agg_netlogo("etox_validation_testing/" + folder + "/out/netlogo.csv", "etox_validation_testing/" + folder + "/netlogo.csv")
        elif agg_net:
            agg_netlogo("etox_validation_testing/" + folder + "/out/netlogo.csv", "etox_validation_testing/" + folder + "/netlogo.csv")
        elif agg_bee:
                agg_beecs("etox_validation_testing/" + folder + "/out/beecs-%04d.csv", "etox_validation_testing/"+ folder +"/beecs.csv")
        plot_quantiles(
            "etox_validation_testing/" + folder + "/netlogo.csv",
            "etox_validation_testing/" + folder + "/beecs.csv",
            "etox_validation_testing/" + folder ,
            "png",
            #"svg",
            appdays[folder],
        )
