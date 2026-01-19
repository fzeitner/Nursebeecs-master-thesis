from os import path

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt


def plot_popstructure(file1, file2, out_dir, format, appday, appdur, multiyear, nurseplot):
    data1 = pd.read_csv(file1, sep=r'\s*;\s*', engine='python')
    data2 = pd.read_csv(file2, sep=r'\s*;\s*', engine='python')

    fig, ax = plt.subplots(figsize=(12, 4))
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

    beec = ax.vlines(-100, 0, 1, color = 'black', linestyle = '-', label = 'beecs')
    nbeec = ax.vlines(-100, 0, 1, color = 'black', linestyle = '--', label = 'Nbeecs')

    # Add the first legend
    #if nurseplot:
        #first_legend = ax.legend([pop, forag, ihb, nurse, larv ], ['TotalPopulation', 'Foragers', 'TotalIHbees', 'Nurses', 'Larvae'], loc='upper right')
    #else:
        #first_legend = ax.legend([pop, forag, ihb, larv ], ['TotalPopulation', 'Foragers', 'TotalIHbees', 'Larvae'], loc='upper right')
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

    #if appday != 0:
        #ax.legend(handles=[beec, nbeec, app], loc='upper left')
    #else:
        #ax.legend(handles=[beec, nbeec], loc='upper left')

    #plt.gca().add_artist(first_legend)
    plt.text(0.05, 0.9, "a", fontsize=20, transform=ax.transAxes, va='top', weight = 'bold')

    if multiyear > 1:
        alignment = 'right'
    else:
        alignment = 'left'
    size = str(12-1.5*multiyear)
    ax.set_xticks(xticks, labels, horizontalalignment = alignment, fontsize=size)

    fig.tight_layout()

    plt.savefig(path.join(out_dir, 'PopStructure' + "." + format))
    plt.close()

def plot_multiyear(file1, file2, file3, out_dir, format, appday, appdur, multiyear, nurseplot):
    data1 = pd.read_csv(file1, sep=r'\s*;\s*', engine='python')
    data2 = pd.read_csv(file2, sep=r'\s*;\s*', engine='python')
    data3 = pd.read_csv(file3, sep=r'\s*;\s*', engine='python')

    fig, ax = plt.subplots(figsize=(9, 6))
    #lines = ['-', '--']
        
    #CB_color_cycle = ['#377eb8', '#ff7f00', '#4daf4a',
    #              '#f781bf', '#a65628', '#984ea3',
    #              "#505050", '#e41a1c', '#dede00']
    
    for data, col, model in [
        (data1, "navy", "Nbeecs"),
        (data2, "red", "PFNexp"),
        (data3, "olive", "P_in_red"),
    ]:
        q10 = data["TotalPop" + "_Q05"]
        q90 = data["TotalPop" + "_Q95"]
        q50 = data["TotalPop" + "_Q50"]

        ax.plot(data.ticks, q50, c=col, label=model)
        ax.fill_between(data.ticks, q10, q90, color=col, alpha=0.1)
        ax.plot(data.ticks, data["HoneyEnergyStore_Q50"], c = col, linestyle = "--" )

        #axs[i][j].legend(loc='best')
        #if i == 1 and j == 1:
        #    axs[i][j].legend(loc='lower right')

    ax.set_xlim(0,365*multiyear)
    ax.set_ylim(0, 1.05*max(max(data1["TotalPop"+"_Q95"]), max(data2["TotalPop"+"_Q95"]), max(data3["TotalPop"+"_Q95"])))
    ax.set_ylabel("Total Population Size", fontsize="12")

    #beec = ax.vlines(-100, 0, 1, color = 'black', linestyle = '-', label = 'beecs')
    #nbeec = ax.vlines(-100, 0, 1, color = 'black', linestyle = '--', label = 'Nbeecs')

    # Add the first legend
    #if nurseplot:
        #first_legend = ax.legend([pop, forag, ihb, nurse, larv ], ['TotalPopulation', 'Foragers', 'TotalIHbees', 'Nurses', 'Larvae'], loc='upper right')
    #else:
        #first_legend = ax.legend([pop, forag, ihb, larv ], ['TotalPopulation', 'Foragers', 'TotalIHbees', 'Larvae'], loc='upper right')
    # Add the second legend

    dayspermonth = [31,28,31,30,31,30,31,31,30,31,30,31]
    months = ['F', 'M', 'A', 'M', 'J', 'J', 'A', 'S', 'O', 'N', 'D', 'J']
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
            for i in range(365*2, 365*multiyear, 365):
                ax.axvspan(i+appday, i+appday+appdur, alpha=0.2, color='black', label = 'Application')

    #if appday != 0:
        #ax.legend(handles=[beec, nbeec, app], loc='upper left')
    #else:
        #ax.legend(handles=[beec, nbeec], loc='upper left')

    #plt.gca().add_artist(first_legend)
    #plt.text(0.05, 0.9, "A", fontsize=20, transform=ax.transAxes, va='top', weight = 'bold')

    alignment = 'left'

    #size = str(12-1.5*multiyear)
    ax.set_xticks(range(0, 365*multiyear, 365), labels= range(1,multiyear+1,1), size = 12, )
    ax.set_xticks(xticks, minor=True, )#labels = labels, horizontalalignment = alignment, size = 6)

    fig.tight_layout()

    plt.savefig(path.join(out_dir, 'MultiyearPlot' + "." + format))
    plt.close()

def plot_adultstructure(file1, file2, out_dir, format, multiyear):
    data1 = pd.read_csv(file1, sep=r'\s*;\s*', engine='python')
    data2 = pd.read_csv(file2, sep=r'\s*;\s*', engine='python')

    fig, ax = plt.subplots(nrows= 1, ncols=2, sharey=True, figsize=(12, 3))
        
    CB_color_cycle = ['#377eb8', '#ff7f00', '#4daf4a',
                  '#f781bf', '#a65628', '#984ea3',
                  "#505050", '#e41a1c', '#dede00']
    alpha = 1.0
    
    #ax[0].plot(data1.ticks, data1['TotalIHbees_Q50'], c= CB_color_cycle[-2])
    ax[0].fill_between(data1.ticks, 0, data1['TotalIHbees_Q50'], color=CB_color_cycle[-2], alpha=alpha)
    #ax[0].plot(data1.ticks, data1['TotalForagers_Q50'] + data1['TotalIHbees_Q50'], c= "navy")
    ax[0].fill_between(data1.ticks, data1['TotalIHbees_Q50'], data1['TotalForagers_Q50'] + data1['TotalIHbees_Q50'], color="navy", alpha=alpha)
    ax[0].set_xlim(0,365)
    """
    #ax[1].plot(data2.ticks, data2['NonNurseIHbees'], c= CB_color_cycle[-2])
    ax[0].fill_between(data1.ticks, 0, data1['NonNurseIHbees'], color=CB_color_cycle[-2], alpha=alpha)
    #ax[1].plot(data2.ticks, data2['NonNurseIHbees'] + data2['IHbeeNurses'], c= CB_color_cycle[1])
    ax[0].fill_between(data1.ticks, data1['NonNurseIHbees'], data1['NonNurseIHbees'] + data1['IHbeeNurses'], color=CB_color_cycle[1], alpha=alpha)
    #ax[1].plot(data2.ticks, data2['NonNurseIHbees'] + data2['IHbeeNurses'] + data2['Winterbees'], c= CB_color_cycle[0])
    ax[0].fill_between(data1.ticks, data1['NonNurseIHbees'] + data1['IHbeeNurses'], data1['NonNurseIHbees'] + data1['IHbeeNurses'] + data1['RevertedForagers'], color=CB_color_cycle[5], alpha=alpha)
    ax[0].fill_between(data1.ticks, data1['NonNurseIHbees'] + data1['IHbeeNurses'] + data1['RevertedForagers'], data1['NonNurseIHbees'] + data1['IHbeeNurses'] + data1['RevertedForagers'] + data1['Winterbees'], color=CB_color_cycle[0], alpha=alpha)
    #ax[1].plot(data2.ticks, data2['NonNurseIHbees'] + data2['IHbeeNurses'] + data2['Winterbees'] + data2['NonWinterbees'], c="navy")
    ax[0].fill_between(data1.ticks, data1['NonNurseIHbees'] + data1['IHbeeNurses'] + data1['RevertedForagers'] + data1['Winterbees'], data1['NonNurseIHbees'] + data1['IHbeeNurses'] + data1['RevertedForagers'] + data1['Winterbees'] + data1['NormalForagers'] , color="navy", alpha=alpha)
    #ax[0].fill_between(data1.ticks, data1['NonNurseIHbees'] + data1['IHbeeNurses'] + data1['RevertedForagers'] + data1['Winterbees'], 100, color="navy", alpha=alpha)
    ax[0].set_xlim(0,365)
    """ 
    #ax[1].plot(data2.ticks, data2['NonNurseIHbees_Q50'], c= CB_color_cycle[-2])
    ax[1].fill_between(data2.ticks, 0, data2['NonNurseIHbees_Q50'], color=CB_color_cycle[-2], alpha=alpha)
    #ax[1].plot(data2.ticks, data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'], c= CB_color_cycle[1])
    ax[1].fill_between(data2.ticks, data2['NonNurseIHbees_Q50'], data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'], color=CB_color_cycle[1], alpha=alpha)
    #ax[1].plot(data2.ticks, data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'] + data2['Winterbees_Q50'], c= CB_color_cycle[0])
    ax[1].fill_between(data2.ticks, data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'], data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'] + data2['RevertedForagers_Q50'], color=CB_color_cycle[5], alpha=alpha)
    ax[1].fill_between(data2.ticks, data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'] + data2['RevertedForagers_Q50'], data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'] + data2['RevertedForagers_Q50'] + data2['Winterbees_Q50'], color=CB_color_cycle[0], alpha=alpha)
    #ax[1].plot(data2.ticks, data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'] + data2['Winterbees_Q50'] + data2['NonWinterbees_Q50'], c="navy")
    #if data2['NonNurseIHbees_Q50'].empty() + data2['IHbeeNurses_Q50'] + data2['RevertedForagers_Q50'] + data2['Winterbees_Q50'] +data2['NormalForagers_Q50'] == 0.:
    #ax[1].fill_between(data2.ticks, data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'] + data2['RevertedForagers_Q50'] + data2['Winterbees_Q50'], data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'] + data2['RevertedForagers_Q50'] + data2['Winterbees_Q50'] + data2['NormalForagers_Q50'] , color="navy", alpha=alpha)
    ax[1].fill_between(data2.ticks, data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'] + data2['RevertedForagers_Q50'] + data2['Winterbees_Q50'], 100, color="navy", alpha=alpha)
    #else:
    #    ax[1].fill_between(data2.ticks, data2['NonNurseIHbees_Q50'] + data2['IHbeeNurses_Q50'] + data2['RevertedForagers_Q50'] + data2['Winterbees_Q50'], 100, color="navy", alpha=alpha)
    ax[1].set_xlim(0,365)

    #ax.set_title('PopStructure')
    ax[0].set_ylabel("Fraction of adult workers [%]", fontsize="12")
    ax[0].set_ylim(0,100)

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

    ax[0].text(0.05, 0.9, "c", fontsize=20, transform=ax[0].transAxes, va='top', weight = 'bold', color = "white")
    ax[1].text(0.05, 0.9, "d", fontsize=20, transform=ax[1].transAxes, va='top', weight = 'bold', color = "white")

    if multiyear > 1:
        alignment = 'right'
    else:
        alignment = 'left'
    size = str(12-1.5*multiyear)
    ax[0].set_xticks(xticks, labels, horizontalalignment = alignment, fontsize=size)
    ax[1].set_xticks(xticks, labels, horizontalalignment = alignment, fontsize=size)

    fig.tight_layout()

    plt.savefig(path.join(out_dir, 'AdultStructure' + "." + format))
    plt.close()


def plot_popmosaic(file1, file2, out_dir, format, appday, appdur):
    data1 = pd.read_csv(file1, sep=r'\s*;\s*', engine='python')
    data2 = pd.read_csv(file2, sep=r'\s*;\s*', engine='python')
    #data3 = pd.read_csv(file3, sep=r'\s*;\s*', engine='python')

    fig, axs = plt.subplots(nrows= 2, ncols=2, sharex=True,  figsize=(12, 6))

    #metrics = ["TotalForagers", "TotalIHbees", "TotalLarvae", "TotalEggs"]
    #metrics = ["ETOX_Mean_Dose_Forager", "ETOX_Mean_Dose_IHbee", "ETOX_Mean_Dose_Larvae", "HoneyEnergyStore"]
    metrics = ["TotalPop", "Aff", "TotalForagers", "TotalIHbees"]
    #metrics = ["Aff", "NurseAgeMax", "NurseWorkLoad", "ProteinFactorNurses"]
    labels = metrics
    #labels = ["Mean Dose per Forager [µg]", "Mean Dose per IHbee [µg]", "Mean Dose per Larvae [µg]", "Honey Store [kJ]"]
    #labels = ["TotalPopulation", "WorkerLarvae", "NurseWorkload [-]", "ProteinFactorNurses [-]"]
    
    for i in range(2):
        for j in range(2):
            for data, col, model in [
                (data1, "navy", "beecs"),
                (data2, "red", "Nbeecs"),
                #(data3, "olive", "Cannibalism"),
            ]:
                q10 = data[metrics[i*2+j] + "_Q05"]
                q90 = data[metrics[i*2+j] + "_Q95"]
                q50 = data[metrics[i*2+j] + "_Q50"]

                axs[i][j].plot(data.ticks, q50, c=col, label=model)
                axs[i][j].fill_between(data.ticks, q10, q90, color=col, alpha=0.1)
                #if i == 0 and j == 1:
                    #axs[i][j].plot(data.ticks, data["TotalNurses_Q50"], c=col, label=model, linestyle = "--")

            axs[i][j].set_ylim(0, 1.05*max(max(data1[metrics[i*2+j]+"_Q95"]), max(data2[metrics[i*2+j]+"_Q95"])))
            axs[i][j].set_ylabel(labels[i*2+j], fontsize="12")
            if appday > 0:
                axs[i][j].axvspan(appday, appday+appdur, alpha=0.3, color='black', label = 'Application')
            #axs[i][j].legend(loc='best')
            #if i == 1 and j == 1:
            #    axs[i][j].legend(loc='lower right')
    
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



def plot_larvaratios(file1, file2, file3, file4, out_dir, format):
    data1 = pd.read_csv(file1, sep=r'\s*;\s*', engine='python')
    data2 = pd.read_csv(file2, sep=r'\s*;\s*', engine='python')
    data3 = pd.read_csv(file3, sep=r'\s*;\s*', engine='python')
    data4 = pd.read_csv(file4, sep=r'\s*;\s*', engine='python')

    fig, axs = plt.subplots(nrows= 1, ncols=2, sharex=True,  figsize=(12, 4))

    metrics = ["WorkerLarvaRatio",]
    labels = metrics
    for j, data, col, model in [
                (0, data1, "navy", "beecs"),
                (0, data2, "red", "Nbeecs"),
                (1, data3, "navy", "beecs"),
                (1, data4, "red", "Nbeecs"),          
    ]:

        q10 = data["WorkerLarvaRatio" + "_Q05"]
        q90 = data["WorkerLarvaRatio" + "_Q95"]
        q50 = data["WorkerLarvaRatio" + "_Q50"]

        axs[j].plot(data.ticks, q50, c=col, label=model)
        axs[j].fill_between(data.ticks, q10, q90, color=col, alpha=0.1)
        axs[j].plot(data.ticks, data["FractionNurses_Q50"], c=col, label=model, linestyle = "--")
        axs[j].hlines(0.5,0,365,linestyle = ":", color = "black")
        if j == 0:
            axs[0].text(0.05, 0.9, "A", fontsize=15, transform=axs[0].transAxes, va='top', weight = 'bold', color = "black")
        if j == 1:
            axs[1].text(0.05, 0.9, "B", fontsize=15, transform=axs[1].transAxes, va='top', weight = 'bold', color = "black")
    
    dayspermonth = [31,28,31,30,31,30,31,31,30,31,30,31]
    months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    labels = months
    xticks = [0]
    for i in range(11):
        xticks.append(xticks[-1]+dayspermonth[i])

    axs[0].set_xlim(0,365)
    axs[0].set_xticks(xticks, labels, horizontalalignment = "left", fontsize="10")

    fig.tight_layout()

    plt.savefig(path.join(out_dir, 'Larvaratios' + "." + format))
    plt.close()




def plot_6mosaic(file1, file2, out_dir, format, appday, appdur):
    data1 = pd.read_csv(file1, sep=r'\s*;\s*', engine='python')
    data2 = pd.read_csv(file2, sep=r'\s*;\s*', engine='python')
    #data3 = pd.read_csv(file3, sep=r'\s*;\s*', engine='python')

    fig, axs = plt.subplots(nrows= 3, ncols=2, sharex=True,  figsize=(12, 9))

    #metrics = ["TotalForagers", "TotalIHbees", "Aff", "NurseAgeMax", "NurseWorkLoad", "ProteinFactorNurses"] 
    #metrics = ["TotalForagers", "ETOX_Mean_Dose_Forager_mug", "TotalIHbees", "ETOX_Mean_Dose_IHbee_mug", "TotalLarvae", "ETOX_Mean_Dose_Larvae_mug", "TotalNurses", "ETOX_Mean_Dose_Nurses_mug" ] 
    metrics = ["Aff", "NurseAgeMax", "ProteinFactorNurses", "NurseWorkLoad", "NurseMaxPollenIntake", "NurseMaxHoneyIntake",] 
    #labels = ["TotalForagers", "TotalIHbees", "AFF [d]", "NurseAgeMax [d]", "NurseWorkload", "ProteinFactorNurses",]
    #labels = ["TotalForagers", "Mean Dose per Forager [µg]", "TotalIHbees", "Mean Dose per IHbee [µg]", "TotalLarvae", "Mean Dose per Larva [µg]" ]
    labels = ["AFF [d]", "NurseAgeMax [d]", "ProteinFactorNurses", "NurseWorkLoad", "NursePollenIntake [mg]", "NurseHoneyIntake [mg]" ]
    
    for i in range(3):
        for j in range(2):
            for data, col, model in [
                (data1, "navy", "beecs"),
                (data2, "red", "Nbeecs"),
                #(data2, "olive", "Cannibalism"),
            ]:
                q10 = data[metrics[i*2+j] + "_Q05"]
                q90 = data[metrics[i*2+j] + "_Q95"]
                q50 = data[metrics[i*2+j] + "_Q50"]

                axs[i][j].plot(data.ticks, q50, c=col, label=model)
                axs[i][j].fill_between(data.ticks, q10, q90, color=col, alpha=0.1)

                #if i == 1:
                #    q50 = data[metrics[(i+2)*2+j] + "_Q50"]
                #    axs[i][j].plot(data.ticks, q50, c=col, linestyle = ":", label=model)
                #    axs[i][j].fill_between(data.ticks, q10, q90, color=col, alpha=0.1)
                #if i*2 + j == 3:
                #    axs[i][j].hlines(1.0, 0, 365, linestyle = "--", color = "black", linewidth = 1)
                #    axs[i][j].hlines(1.5, 0, 365, linestyle = "--", color = "black", linewidth = 1)

            axs[i][j].set_ylim(0, 1.05*max(max(data1[metrics[i*2+j]+"_Q95"]), max(data2[metrics[i*2+j]+"_Q95"])))#, max(data3[metrics[i*2+j]+"_Q95"])))
            #if j == 1:
            #    axs[i][j].set_ylim(0, 1.05*max(max(data1[metrics[i*2+j]+"_Q50"]*1.5), max(data2[metrics[i*2+j]+"_Q50"]*1.5)))

            axs[i][j].set_ylabel(labels[i*2+j], fontsize="12")
            if appday > 0:
                axs[i][j].axvspan(appday, appday+appdur, alpha=0.3, color='black', label = 'Application')
            #axs[i][j].legend(loc='best')
            #if i == 1 and j == 1:
            #    axs[i][j].legend(loc='lower right')
    
    dayspermonth = [31,28,31,30,31,30,31,31,30,31,30,31]
    months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    labels = months
    xticks = [0]
    for i in range(11):
        xticks.append(xticks[-1]+dayspermonth[i])

    axs[0][0].set_xlim(0,365)
    axs[0][0].set_xticks(xticks, labels, horizontalalignment = "left", fontsize="10")

    fig.tight_layout()

    plt.savefig(path.join(out_dir, '6Mosaic' + "." + format))
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
            "Rothamsted2009_clothianidin_5years": 8,
    }
    appdur = {"default_beecs" : 0,  
            "default_etox" : 0,                     
            "default_dimethoate": 9, 
            "Rothamsted2009_fenoxycarb": 8, 
            "Rothamsted2009_etox": 0,
            "Rothamsted2009_beecs": 0,
            "Rothamsted2009_fenoxycarb_5years" : 8,
            "Rothamsted2009_etox_5years": 0,
            "Rothamsted2009_clothianidin_5years": 30,
    }

    testfolders = ["default_etox", "default_dimethoate", "default_beecs", "Rothamsted2009_beecs",
                   "Rothamsted2009_fenoxycarb", "Rothamsted2009_etox", "Rothamsted2009_fenoxycarb_5years", "Rothamsted2009_etox_5years",  "Rothamsted2009_clothianidin_5years",]

    folder = testfolders[0]
    fileformats = ["svg", "png"]
    format = fileformats[0]

    plotall = False
    plot_nursebees = False


    if plotall:
        for folder in testfolders:
            plot_popmosaic(
                    "nursebeecs_testing/" + folder + "/beecs.csv",
                    "nursebeecs_testing/" + folder + "/new.csv",
                    "nursebeecs_testing/" + folder ,
                    format,
                    appdays[folder],
                    appdur[folder],
            )
            plot_popstructure(   
                    "nursebeecs_testing/" + folder + "/beecs.csv",
                    "nursebeecs_testing/" + folder + "/new.csv",
                    "nursebeecs_testing/" + folder ,
                    format,
                    appdays[folder],
                    appdur[folder],
                    multiyear_app[folder],
                    plot_nursebees,
            )
    else:
        
        plot_popmosaic(
                    "nursebeecs_testing/" + folder + "/beecs.csv",
                    #"nursebeecs_testing/" + folder + "/old.csv",
                    "nursebeecs_testing/" + folder + "/new.csv",
                    "nursebeecs_testing/" + folder ,
                    format,
                    appdays[folder],
                    appdur[folder],
        )
        plot_popstructure(   
                "nursebeecs_testing/" + folder + "/beecs.csv",
                "nursebeecs_testing/" + folder + "/new.csv",
                "nursebeecs_testing/" + folder ,
                format,
                appdays[folder],
                appdur[folder],
                multiyear_app[folder],
                plot_nursebees,
        )     
        """
        plot_6mosaic(
                    "nursebeecs_testing/" + folder + "/beecs.csv",
                    #"nursebeecs_testing/" + folder + "/old.csv",
                    "nursebeecs_testing/" + folder + "/new.csv",
                    "nursebeecs_testing/" + folder ,
                    format,
                    appdays[folder],
                    appdur[folder],
        )
        
        plot_adultstructure(   
                "nursebeecs_testing/" + folder + "/beecs.csv",
                "nursebeecs_testing/" + folder + "/new.csv",
                "nursebeecs_testing/" + folder ,
                format,
                multiyear_app[folder],
        ) 
        
        plot_multiyear (
            "nursebeecs_testing/" + folder + "/nbeecs.csv",
            "nursebeecs_testing/" + folder + "/nbeecsHG.csv",
            "nursebeecs_testing/" + folder + "/nbeecsHGFood.csv",
            "nursebeecs_testing/" + folder ,
            format,
            appdays[folder],
            appdur[folder],
            multiyear_app[folder],
            False
            )
        plot_larvaratios(
            "nursebeecs_testing/" + testfolders[0] + "/beecs.csv",
            "nursebeecs_testing/" + testfolders[0] + "/new.csv",
            "nursebeecs_testing/" + testfolders[5] + "/beecs.csv",
            "nursebeecs_testing/" + testfolders[5] + "/new.csv",
            "nursebeecs_testing/" + testfolders[0] ,
            format,

        )"""

