# WotWizard

WotWizard builds, from the blockchain and sandbox of Duniter, a prediction of the future entries of candidates into the Duniter Web of Trust (WOT). It uses a simulation of the Duniter mechanism. When several possibilities may happen, each one is listed with its probabilities. The published lists are automatically updated every five minutes, and any change is signalled visually.

Several other tools are provided, such as a fast "Web of Trust" Explorer.

This program needs that a Duniter node runs on the same computer.

There are two parts in WotWizard: first a server (wwServer), written in Go (v1.13.6) under GNU / Linux and which communicates through files containing, on input, GraphQL requests, and on output, JSON answers; second a graphical user interface (WotWizard.exe) written in Component Pascal (BlackBox v1.7.1 under Wine), which generates GraphQL requests through menus and interactive windows, and display answers with texts or graphics.

You can find the executables here:
	https://github.com/duniter/WotWizard/releases

WotWizard may be used as well as a server, or as a stand-alone application with a front-end.

The version of the associated Duniter node must be 1.7.17 at least.
