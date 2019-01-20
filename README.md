# WotWizard

WotWizard builds, from the blockchain and sandbox of Duniter, a prediction of the future entries of candidates into the Duniter Web of Trust (WOT). It uses a simulation of the Duniter mechanism. When several possibilities may happen, each one is listed with its probabilities. The published lists are automatically updated every five minutes, and any change is signalled visually.

This program needs that a Duniter node runs on the same computer.
This program runs natively on Windows. If your computer runs on Linux, install "wine" first, and, in a terminal, run "winecfg" and configure the drives to be sure that the Duniter database (see below) can be reached.

You can find the executables here:
	https://github.com/duniter/WotWizard/releases

There are two versions of WotWizard: Server or Standalone.

**Warning**: You should use the 1.6.29 version of duniter or any later version in the series 1.6.x. Don't use a 1.7.x version.

Server version:

Server version is divided into:
    - a server (WWServer.exe) which produces two data files (WWByDates.json & WWMeta.json)
    - a gui interface (WotWizard.exe) which reads the data from WWServer.exe and displays them

You can run this version as a server only, if you want to display yourself its results (e.g. in a web page). In this case, put the file "WWServer.exe" into an empty directory and put the Windows dll "sqlite3.dll" into the same directory. Then, you have to run "WWServer.exe" as often as you want with the line command:
	
	$wine WWServer.exe
	
or directly from within your application. The dll can be found at the address:
	
	https://www.sqlite.org/2015/sqlite-dll-win32-x86-3081002.zip

At the end of every run, WWServer writes (or updates) two files, in json syntax: WWByDates.json, which lists the entries of the newcomers, sorted and grouped by dates, and WWMeta.json, which gives some metadata on the computing and the original data.

If you want to display the updated results of WWServer continuously, put the file "WotWizard.exe" into the same directory as "WWServer.exe" and run WotWizard with the line command:
	
	$wine WotWizard.exe

You can:
	- choose the way the list is displayed (by names or by dates, or metadata)
	- manually update the list (it's automatically updated every five minutes)



Standalone version:

This version combines the server side and the gui side into a single application. It contains several derived tools like an explorer of the web of trust, a tool that prints the number of members day by day, or the number of certifications by certifiers, etc... Its installation is the same as for the Server version; just replace WWServer.exe by WotWizard.exe (Standalone Version).
