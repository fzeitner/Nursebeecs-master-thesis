from os import path

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt


def plot_popstructure(file1, file2, out_dir, format, appday, appdur, multiyear, nurseplot):
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

    #ax.set_title('PopStructure')
    ax.set_ylabel("Individuals [-]", fontsize="12")
    ax.set_xlim(0,365*multiyear)
    ax.set_ylim(0,max(max(data1['TotalPop_Q95']), max(data2['TotalPop_Q95'])))

    beec = ax.vlines(-100, 0, 1, color = 'black', linestyle = '-', label = 'Netlogo')
    nbeec = ax.vlines(-100, 0, 1, color = 'black', linestyle = '--', label = 'beecs')

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
            app = ax.axvspan(appday, appday+appdur, alpha=0.1, color='black', label = 'Application')
            for i in range(1,multiyear):
                ax.axvspan(appday, appday+appdur, alpha=0.1, color='black', label = 'Application')
    elif appday > 0:
        app = ax.axvspan(appday, appday+appdur, alpha=0.3, color='black', label = 'Application')

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


def plot_popmosaic(file1, file2, out_dir, format, appday, appdur):
    data1 = pd.read_csv(file1, sep=r'\s*;\s*', engine='python')
    data2 = pd.read_csv(file2, sep=r'\s*;\s*', engine='python')

    fig, axs = plt.subplots(nrows= 2, ncols=2, sharex=True,  figsize=(12, 6))

    metrics = ["TotalForagers", "TotalIHbees", "TotalLarvae", "TotalEggs"]
    #metrics = ["ETOX_Mean_Dose_Forager", "ETOX_Mean_Dose_IHbee", "ETOX_Mean_Dose_Larvae", "HoneyEnergyStore"]
    labels = metrics
    #labels = ["Mean Dose per Forager [µg]", "Mean Dose per IHbee [µg]", "Mean Dose per Larvae [µg]", "Honey Store [kJ]"]
    
    for i in range(2):
        for j in range(2):
            for data, col, model in [
                (data1, "navy", "Netlogo"),
                (data2, "red", "beecs"),
            ]:
                q10 = data[metrics[i*2+j] + "_Q05"]
                q90 = data[metrics[i*2+j] + "_Q95"]
                q50 = data[metrics[i*2+j] + "_Q50"]

                axs[i][j].plot(data.ticks, q50, c=col, label=model)
                axs[i][j].fill_between(data.ticks, q10, q90, color=col, alpha=0.1)

            axs[i][j].set_ylim(0, 1.05*max(max(data1[metrics[i*2+j]+"_Q95"]), max(data2[metrics[i*2+j]+"_Q95"])))
            axs[i][j].set_ylabel(labels[i*2+j], fontsize="12")
            if appday > 0:
                axs[i][j].axvspan(appday, appday+appdur, alpha=0.3, color='black', label = 'Application')
            axs[i][j].legend(loc='best')
    
    dayspermonth = [31,28,31,30,31,30,31,31,30,31,30,31]
    months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    labels = months
    xticks = [0]
    for i in range(11):
        xticks.append(xticks[-1]+dayspermonth[i])

    axs[0][0].set_xlim(0,365)
    axs[0][0].set_xticks(xticks, labels, horizontalalignment = "left", fontsize="10")

    fig.tight_layout()

    plt.savefig(path.join(out_dir, 'PopMosaic' + "." + format))
    plt.close()


if __name__ == "__main__":
    ### change test folder and day of application manually here, applicationday is only relevant for 
    ### adding a visual indicator in plots, does not change anything regarding the results
    appdays = {"default_etox" : 0,                     # appday = 0 for no application
              "default_dimethoate": 217, 
              "Rothamsted2009_fenoxycarb": 189, 
              "Rothamsted2009_etox": 0,
              "default_beecs": 0, 
              "Rothamsted2009_beecs": 0,
    }
    appdur = {"default_etox" : 0,                     # appday = 0 for no application
              "default_dimethoate": 9, 
              "Rothamsted2009_fenoxycarb": 8, 
              "Rothamsted2009_etox": 0,
              "default_beecs": 0, 
              "Rothamsted2009_beecs": 0,
    }

    testfolders = ["default_etox", "default_dimethoate", "Rothamsted2009_fenoxycarb", "Rothamsted2009_etox", 
                   "default_beecs", "Rothamsted2009_beecs"]
    folder = testfolders[2]
    fileformats = ["svg", "png"]
    format = fileformats[0]

    plotall = False


    if plotall:
        for folder in testfolders:

            plot_popmosaic(
                    "etox_validation_testing/" + folder + "/netlogo.csv",
                    "etox_validation_testing/" + folder + "/beecs.csv",
                    "etox_validation_testing/" + folder ,
                    format,
                    appdays[folder],
                    appdur[folder],
            )
            plot_popstructure(   
                    "etox_validation_testing/" + folder + "/netlogo.csv",
                    "etox_validation_testing/" + folder + "/beecs.csv",
                    "etox_validation_testing/" + folder ,
                    format,
                    appdays[folder],
                    appdur[folder],
                    1,
                    False,
            )
    else:
        plot_popmosaic(
                    "etox_validation_testing/" + folder + "/netlogo.csv",
                    "etox_validation_testing/" + folder + "/beecs.csv",
                    "etox_validation_testing/" + folder ,
                    format,
                    appdays[folder],
                    appdur[folder],
        )
        plot_popstructure(   
                "etox_validation_testing/" + folder + "/netlogo.csv",
                "etox_validation_testing/" + folder + "/beecs.csv",
                "etox_validation_testing/" + folder ,
                format,
                appdays[folder],
                appdur[folder],
                1,
                False,
        )


