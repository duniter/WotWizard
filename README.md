# WotWizard

WotWizard builds, from the blockchain and sandbox of Duniter, a prediction of the future entries of candidates into the Duniter Web of Trust (WOT). It uses a simulation of the Duniter mechanism. When several possibilities may happen, each one is listed with its probabilities. The published lists are automatically updated every five minutes, and any change is signalled visually.

Several other tools are provided, such as a fast "Web of Trust" Explorer.

This program needs that a Duniter node runs on the same computer.

WotWizard includes a server executable (wwServer) written in Go (v1.15.2) and running under GNU / Linux. This server communicates by the way of http POST methods containing GraphQL requests on input, and JSON answers on output. WotWizard includes an optional graphical user interface (wwClient) written in Go too and interacting with the server through http and with users through a web browser.

You can find the executables here:

  https://github.com/duniter/WotWizard/releases

The version of the associated Duniter node must be 1.7.17 at least. Versions 1.8.x don't work with WotWizard.

All included softwares have a GPLv3 license.

The graphQL type system definition text for the WotWizard server can be found in Help/Typesystem.txt.

