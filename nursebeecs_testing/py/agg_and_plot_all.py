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

            #q = np.nanpercentile(values, [0.05, 0.1, 0.25, 0.5, 0.75, 0.9, 0.95])
            q = np.quantile(values, [0.05, 0.1, 0.25, 0.5, 0.75, 0.9, 0.95])
            out.loc[tick, cols] = q
        out = out.copy()      # to keep df from becoming highly fragmented

    out.to_csv(out_file, sep=";", index=False)

def agg_beecs_mean(file_pattern, out_file):
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
        for tick in ticks:
            values = data[column][data.ticks == tick]

            inf = np.isinf(values)
            if sum(inf) > 0: # debugging anchor
                _ = 1

            q = np.mean(values)
            out.loc[tick, column] = q
        out = out.copy()      # to keep df from becoming highly fragmented

    out.to_csv(out_file, sep=";", index=False)


def plot_quantiles(beecs_file, nbeecs_file, nbeecs2_file, out_dir, format, appday, multiyear):
    data_nbeecs = pd.read_csv(nbeecs_file, sep=r'\s*;\s*', engine='python')
    data_beecs = pd.read_csv(beecs_file, sep=r'\s*;\s*', engine='python')
    data_nbeecs2 = pd.read_csv(nbeecs2_file, sep=r'\s*;\s*', engine='python')

    columns = list(data_nbeecs.columns)[1:]
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
            data_beecs,
            data_nbeecs,
            data_nbeecs2, 
            col,
            quantiles,
            path.join(out_dir, col + "." + format),
            appday,
            multiyear
        )


def plot_column(data_beecs, data_nbeecs, data_nbeecs2, column, quantiles, image_file, appday, multiyear):
    median_col = quantiles[len(quantiles) // 2][0]

    fig, ax = plt.subplots(figsize=(10, 4))
    for data, col, model in [
        (data_beecs, "blue", "nbeecs"),
        (data_nbeecs, "red", "nbeecsHG"),
        (data_nbeecs2, "green", "nbeecsHGFood"),
    ]:      
        q10 = data[column + "_Q05"]
        q90 = data[column + "_Q95"]
        q50 = data[column + "_Q50"]

        ax.plot(data.ticks, q50, c=col, label=model)
        ax.fill_between(data.ticks, q10, q90, color=col, alpha=0.1)

    ax.set_title(column)
    #ax.set_xlabel("month", fontsize="12")
    ax.set_xlim(0,365*multiyear)

    dayspermonth = [31,28,31,30,31,30,31,31,30,31,30,31]
    months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    labels = months
    xticks = [0]
    for i in range(11):
        xticks.append(xticks[-1]+dayspermonth[i])

    if multiyear > 1:
        xticks = [dayspermonth[0]]
        for i in range(12*multiyear-1):
            xticks.append(xticks[-1]+dayspermonth[i%12])
        labels = multiyear * months
        if appday != 0:
            ax.vlines(appday+365, 0, max(q90), linestyle = "--", color = "gray", label = "application day")   # have to change the appday manually in func
            for i in range(1,multiyear):
                ax.vlines(appday+i*365, 0, max(q90), linestyle = "--", color = "gray")   # have to change the appday manually in func
    elif appday > 0:
        ax.vlines(appday, 0, max(q90), linestyle = "--", color = "gray", label = "application day")   # have to change the appday manually in func

    if multiyear > 1:
        alignment = 'right'
    else:
        alignment = 'left'
    size = str(12-1.5*multiyear)
    ax.set_xticks(xticks, labels, horizontalalignment = alignment, fontsize=size)

    ax.legend()
    fig.tight_layout()

    plt.savefig(image_file)
    plt.close()


def plot_popstructure(file1, file2, out_dir, format, appday, multiyear, nurseplot):
    data1 = pd.read_csv(file1, sep=r'\s*;\s*', engine='python')
    data2 = pd.read_csv(file2, sep=r'\s*;\s*', engine='python')

    fig, ax = plt.subplots(figsize=(10, 4))
    lines = ['-', '--']
        
    CB_color_cycle = ['#377eb8', '#ff7f00', '#4daf4a',
                  '#f781bf', '#a65628', '#984ea3',
                  "#505050", '#e41a1c', '#dede00']


    pop, = ax.plot(data1.ticks, data1['TotalPop_Q50'], c= 'black', linestyle = lines[0], label='TotalPopulation')
    forag, = ax.plot(data1.ticks, data1['TotalForagers_Q50'], c= CB_color_cycle[0], linestyle = lines[0], label='Foragers')
    ihb, = ax.plot(data1.ticks, data1['TotalIHbees_Q50'], c= CB_color_cycle[7], linestyle = lines[0], label='TotalIHbees')
    #ax.plot(data1.ticks, data1['TotalPupae_Q50'], c= 'yellow', linestyle = lines[0], label='Pupae')
    larv, = ax.plot(data1.ticks, data1['TotalLarvae_Q50'], c= CB_color_cycle[2], linestyle = lines[0], label='Larvae')
    #ax.plot(data1.ticks, data1['TotalEggs_Q50'], c= 'gray', linestyle = lines[0], label='Eggs')
    if nurseplot:
        nurse, = ax.plot(data1.ticks, data1['TotalNurses_Q50'], c= CB_color_cycle[1], linestyle = lines[0], label='Nurses')

    ax.plot(data2.ticks, data2['TotalPop_Q50'], c= 'black', linestyle = lines[1])
    ax.plot(data2.ticks, data2['TotalForagers_Q50'], c= CB_color_cycle[0], linestyle = lines[1])
    ax.plot(data2.ticks, data2['TotalIHbees_Q50'], c= CB_color_cycle[7], linestyle = lines[1])
    #ax.plot(data2.ticks, data2['TotalPupae_Q50'], c= 'yellow', linestyle = lines[1])
    ax.plot(data2.ticks, data2['TotalLarvae_Q50'], c= CB_color_cycle[2], linestyle = lines[1])
    #ax.plot(data2.ticks, data2['TotalEggs_Q50'], c= 'gray', linestyle = lines[1])
    if nurseplot:
        ax.plot(data2.ticks, data2['TotalNurses_Q50'], c= CB_color_cycle[1], linestyle = lines[1], label='Nurses')

    ax.set_title('PopStructure')
    ax.set_ylabel("Individuals [-]", fontsize="12")
    ax.set_xlim(0,365*multiyear)

    beec = ax.vlines(-100, 0, 1, color = 'black', linestyle = '-', label = 'beecs')
    nbeec = ax.vlines(-100, 0, 1, color = 'black', linestyle = '--', label = 'ConumptionRework')

    # Add the first legend
    if nurseplot:
        first_legend = ax.legend([pop, forag, ihb, nurse, larv ], ['TotalPopulation', 'Foragers', 'TotalIHbees', 'Nurses', 'Larvae'], loc='upper right')
    else:
        first_legend = ax.legend([pop, forag, ihb, larv ], ['TotalPopulation', 'Foragers', 'TotalIHbees', 'Larvae'], loc='upper right')
    # Add the second legend

    dayspermonth = [31,28,31,30,31,30,31,31,30,31,30,31]
    months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    labels = months
    xticks = [0]
    for i in range(11):
        xticks.append(xticks[-1]+dayspermonth[i])

    if multiyear > 1:
        xticks = [dayspermonth[0]]
        for i in range(12*multiyear-1):
            xticks.append(xticks[-1]+dayspermonth[i%12])
        labels = multiyear * months
        if appday != 0:
            app = ax.vlines(appday+365, 0, max(max(data1['TotalPop_Q50']), max(data2['TotalPop_Q50'])), linestyle = "-.", color = CB_color_cycle[6], linewidth = 2., label = "application day")   # have to change the appday manually in func
            for i in range(1,multiyear):
                ax.vlines(appday+i*365, 0, max(max(data1['TotalPop_Q50']), max(data2['TotalPop_Q50'])), linestyle = "-.", color = CB_color_cycle[6], linewidth = 2.,)   # have to change the appday manually in func
    elif appday > 0:
        app = ax.vlines(appday, 0, max(max(data1['TotalPop_Q50']), max(data2['TotalPop_Q50'])), linestyle = "-.", color = CB_color_cycle[6], linewidth = 2., label = "application day")   # have to change the appday manually in func

    if appday != 0:
        ax.legend(handles=[beec, nbeec, app], loc='upper left')
    else:
        ax.legend(handles=[beec, nbeec], loc='upper left')

    plt.gca().add_artist(first_legend)

    if multiyear > 1:
        alignment = 'right'
    else:
        alignment = 'left'
    size = str(12-1.5*multiyear)
    ax.set_xticks(xticks, labels, horizontalalignment = alignment, fontsize=size)

    fig.tight_layout()

    plt.savefig(path.join(out_dir, 'PopStructure' + "." + format))
    plt.close()




if __name__ == "__main__":
    ### change test folder and day of application manually here, applicationday is only relevant for 
    ### adding a visual indicator in plots, does not change anything regarding the results
    appdays = {"default_beecs" : 0,  
            "default_etox" : 0,                     # appday = 0 for no application
            "default_dimethoate": 217, 
            "Rothamsted2009_fenoxycarb": 189, 
            "Rothamsted2009_etox": 0,
            "Rothamsted2009_beecs": 0,
            "Rothamsted2009_fenoxycarb_5years" : 189,
            "Rothamsted2009_etox_5years": 0,
            "Rothamsted2009_clothianidin_5years": 182,

    }
    multiyear_app = {"default_beecs" : 1,  
            "default_etox" : 1,                     
            "default_dimethoate": 1, 
            "Rothamsted2009_fenoxycarb": 1, 
            "Rothamsted2009_etox": 1,
            "Rothamsted2009_beecs": 1,
            "Rothamsted2009_fenoxycarb_5years" : 5,
            "Rothamsted2009_etox_5years": 1,
            "Rothamsted2009_clothianidin_5years": 7,
    }

    testfolders = ["default_etox", "default_dimethoate", "default_beecs", "Rothamsted2009_beecs",
                   "Rothamsted2009_fenoxycarb", "Rothamsted2009_etox", "Rothamsted2009_fenoxycarb_5years", "Rothamsted2009_etox_5years",  "Rothamsted2009_clothianidin_5years",]
    file_formats = ["svg", "png"]

    folder = testfolders[0]
    file_format = file_formats[0]       # 0 = svg, 1 = png

    run_all = False                   # True if you want to create all plots at once, just make sure to have run the sims beforehand
    agg_all = True
    agg_nbeecs = False
    agg_beec = False
    plot = False

    if run_all:
        for folder in testfolders:
            if agg_all:
                agg_beecs("nursebeecs_testing/" + folder + "/out/beecs-%04d.csv", "nursebeecs_testing/"+ folder +"/beecs.csv")
                agg_beecs("nursebeecs_testing/" + folder + "/out/oldbc-%04d.csv", "nursebeecs_testing/" + folder + "/oldbc.csv")
                agg_beecs("nursebeecs_testing/" + folder + "/out/newbc-%04d.csv", "nursebeecs_testing/" + folder + "/newbc.csv")
            elif agg_nbeecs:
                agg_beecs("nursebeecs_testing/" + folder + "/out/oldbc-%04d.csv", "nursebeecs_testing/" + folder + "/oldbc.csv")
                agg_beecs("nursebeecs_testing/" + folder + "/out/newbc-%04d.csv", "nursebeecs_testing/" + folder + "/newbc.csv")
            elif agg_beec:
                agg_beecs("nursebeecs_testing/" + folder + "/out/beecs-%04d.csv", "nursebeecs_testing/"+ folder +"/beecs.csv")

            if plot:
                plot_quantiles(
                    "nursebeecs_testing/" + folder + "/beecs.csv",
                    "nursebeecs_testing/" + folder + "/oldbc.csv",
                    "nursebeecs_testing/" + folder + "/newbc.csv",
                    "nursebeecs_testing/" + folder ,
                    file_format,
                    appdays[folder],
                    multiyear_app[folder]
                )
                plot_popstructure(
                    "nursebeecs_testing/" + folder + "/beecs.csv",
                    "nursebeecs_testing/" + folder + "/newbc.csv",
                    "nursebeecs_testing/" + folder ,
                    file_format,
                    appdays[folder],
                    multiyear_app[folder],
                    False
                )
    else:
        if agg_all:
                agg_beecs("nursebeecs_testing/" + folder + "/out/beecs-%04d.csv", "nursebeecs_testing/"+ folder +"/beecs.csv")
                agg_beecs("nursebeecs_testing/" + folder + "/out/oldbc-%04d.csv", "nursebeecs_testing/" + folder + "/oldbc.csv")
                agg_beecs("nursebeecs_testing/" + folder + "/out/newbc-%04d.csv", "nursebeecs_testing/" + folder + "/newbc.csv")
        elif agg_nbeecs:
            agg_beecs("nursebeecs_testing/" + folder + "/out/oldbc-%04d.csv", "nursebeecs_testing/" + folder + "/oldbc.csv")
            agg_beecs("nursebeecs_testing/" + folder + "/out/newbc-%04d.csv", "nursebeecs_testing/" + folder + "/newbc.csv")
        elif agg_beec:
            agg_beecs("nursebeecs_testing/" + folder + "/out/beecs-%04d.csv", "nursebeecs_testing/"+ folder +"/beecs.csv")

        if plot:
            plot_quantiles(
                "nursebeecs_testing/" + folder + "/beecs.csv",
                "nursebeecs_testing/" + folder + "/oldbc.csv",
                "nursebeecs_testing/" + folder + "/newbc.csv",
                "nursebeecs_testing/" + folder ,
                file_format,
                appdays[folder],
                multiyear_app[folder]
            )
            plot_popstructure(
                "nursebeecs_testing/" + folder + "/beecs.csv",
                "nursebeecs_testing/" + folder + "/newbc.csv",
                "nursebeecs_testing/" + folder ,
                file_format,
                appdays[folder],
                multiyear_app[folder],
                False
            )
            """
        if agg_all:
            agg_beecs("nursebeecs_testing/" + folder + "/out/nbeecs-%04d.csv", "nursebeecs_testing/"+ folder +"/nbeecs.csv")
            agg_beecs("nursebeecs_testing/" + folder + "/out/nbeecsHG-%04d.csv", "nursebeecs_testing/" + folder + "/nbeecsHG.csv")
            agg_beecs("nursebeecs_testing/" + folder + "/out/nbeecsHGFood-%04d.csv", "nursebeecs_testing/" + folder + "/nbeecsHGFood.csv")
        elif agg_nbeecs:
            agg_beecs("nursebeecs_testing/" + folder + "/out/nbeecsHG-%04d.csv", "nursebeecs_testing/" + folder + "/nbeecsHG.csv")
            agg_beecs("nursebeecs_testing/" + folder + "/out/nbeecsHGFood-%04d.csv", "nursebeecs_testing/" + folder + "/nbeecsHGFood.csv")
        elif agg_beec:
            agg_beecs("nursebeecs_testing/" + folder + "/out/beecs-%04d.csv", "nursebeecs_testing/"+ folder +"/beecs.csv")
        if plot:
            plot_quantiles(
                "nursebeecs_testing/" + folder + "/nbeecs.csv",
                "nursebeecs_testing/" + folder + "/nbeecsHG.csv",
                "nursebeecs_testing/" + folder + "/nbeecsHGFood.csv",
                "nursebeecs_testing/" + folder ,
                file_format,
                appdays[folder],
                multiyear_app[folder]
            )
            plot_popstructure(
                "nursebeecs_testing/" + folder + "/nbeecs.csv",
                "nursebeecs_testing/" + folder + "/nbeecsHG.csv",
                "nursebeecs_testing/" + folder ,
                file_format,
                appdays[folder],
                multiyear_app[folder],
                False
            )"""

